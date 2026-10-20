#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use std::path::PathBuf;
use std::process::{Child, Command};
use std::sync::Mutex;
use std::thread;
use std::time::Duration;

use tauri::Manager;


#[cfg(windows)]
use std::os::windows::process::CommandExt;

struct BackendHandle {
    child: Mutex<Option<Child>>,
}


fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_updater::Builder::new().build())
        .setup(|app| {
            // Resolve Tauri resource directory.
            let resource_dir = app.path().resource_dir()?;
            let resources_dir = resource_dir.join("resources");

            // Backend executable.
            let backend_path = resources_dir.join("financeapp-backend.exe");

            // eprintln!("DEBUG Tauri resource_dir: {}", resource_dir.display());
            // eprintln!("DEBUG Resources dir: {}", resources_dir.display());
            // eprintln!("DEBUG Backend path: {}", backend_path.display());

            if !backend_path.is_file() {
                return Err(std::io::Error::new(
                    std::io::ErrorKind::NotFound,
                    format!(
                        "Backend executable not found at {}",
                        backend_path.display()
                    ),
                )
                .into());
            }

            let backend_dir = backend_path
                .parent()
                .expect("Backend path has no parent");

            // Processor resources.
            let processor_script = resources_dir
                .join("processor")
                .join("src")
                .join("main.py");

            let requirements_file = resources_dir
                .join("processor")
                .join("requirements-lock.txt");

            let python_bundle = resources_dir
                .join("python-runtime")
                .join("python-3.12.10-amd64.zip");

            let wheels_path = resources_dir.join("python-wheels");

            // Resolve Python from the marker created by prepare-python.ps1.
            // The virtual environment directory includes a dynamic requirements hash.
            let program_data = std::env::var("PROGRAMDATA")
                .unwrap_or_else(|_| r"C:\ProgramData".to_string());

            let marker_path = PathBuf::from(program_data)
                .join("FinanceApp")
                .join(".python-ready");

            let marker_content = std::fs::read_to_string(&marker_path)
                .map_err(|e| {
                    std::io::Error::new(
                        std::io::ErrorKind::NotFound,
                        format!(
                            "Could not read Python environment marker at {}: {}. \
                             Run the installer again to prepare Python.",
                            marker_path.display(),
                            e
                        ),
                    )
                })?;

            let venv_path = marker_content
                .lines()
                .find_map(|line| line.strip_prefix("venv_path="))
                .map(str::trim)
                .filter(|path| !path.is_empty())
                .map(PathBuf::from)
                .ok_or_else(|| {
                    std::io::Error::new(
                        std::io::ErrorKind::InvalidData,
                        format!(
                            "The Python marker does not contain a valid venv_path: {}",
                            marker_path.display()
                        ),
                    )
                })?;

            let python_path = venv_path
                .join("Scripts")
                .join("python.exe");

            // Validate required processor resources.
            if !processor_script.is_file() {
                return Err(std::io::Error::new(
                    std::io::ErrorKind::NotFound,
                    format!(
                        "Processor script not found at {}",
                        processor_script.display()
                    ),
                )
                .into());
            }

            if !requirements_file.is_file() {
                return Err(std::io::Error::new(
                    std::io::ErrorKind::NotFound,
                    format!(
                        "Processor requirements file not found at {}",
                        requirements_file.display()
                    ),
                )
                .into());
            }

            if !python_path.is_file() {
                return Err(std::io::Error::new(
                    std::io::ErrorKind::NotFound,
                    format!(
                        "Python executable from marker not found at {}. \
                         Check the installation log and rerun the installer if necessary.",
                        python_path.display()
                    ),
                )
                .into());
            }

            // eprintln!("DEBUG Processor script: {}", processor_script.display());
            // eprintln!("DEBUG Processor requirements: {}", requirements_file.display());
            // eprintln!("DEBUG Python marker: {}", marker_path.display());
            // eprintln!("DEBUG Python executable: {}", python_path.display());

            let mut cmd = Command::new(&backend_path);
            cmd.current_dir(backend_dir);

            // Configure per-user application data paths.
            if let Ok(appdata) = std::env::var("APPDATA") {
                let base = PathBuf::from(appdata).join("FinanceApp");

                let db_path = base
                    .join("data")
                    .join("db")
                    .join("finance.db");

                let artifacts_path = base
                    .join("data")
                    .join("artifacts");

                let logs_path = base.join("logs");

                std::fs::create_dir_all(
                    db_path.parent().expect("Database path has no parent"),
                )?;
                std::fs::create_dir_all(&artifacts_path)?;
                std::fs::create_dir_all(&logs_path)?;

                cmd.env("DATABASE_PATH", &db_path);
                cmd.env("ARTIFACTS_PATH", &artifacts_path);
                cmd.env("LOGS_PATH", &logs_path);
            }

            // Configure backend and processor.
            cmd.env("PORT", "8081");
            cmd.env("PROCESSOR_SCRIPT", &processor_script);
            cmd.env("PROCESSOR_REQUIREMENTS", &requirements_file);
            cmd.env("PROCESSOR_PYTHON", &python_path);

            if python_bundle.is_file() {
                cmd.env("PYTHON_BUNDLE_PATH", &python_bundle);
            }

            if wheels_path.is_dir() {
                cmd.env("PYTHON_WHEELS_PATH", &wheels_path);
            }

            // Prevent a separate console window on Windows.
            #[cfg(windows)]
            {
                cmd.creation_flags(0x08000000); // CREATE_NO_WINDOW
            }

            // Start backend.
            let mut child: Child = cmd.spawn().map_err(|e| {
                // eprintln!("Failed to start backend: {}", e);
                std::io::Error::other(format!(
                    "Failed to start backend at {}: {}",
                    backend_path.display(),
                    e
                ))
            })?;

            // eprintln!("Backend started with PID {}", child.id());

            // Wait for the backend health endpoint.
            let health_url = "http://127.0.0.1:8081/health";
            let mut attempts = 0;
            let max_attempts = 60;

            loop {
                attempts += 1;

                if let Ok(resp) = reqwest::blocking::get(health_url) {
                    if resp.status().is_success() {
                        break;
                    }
                }

                if attempts >= max_attempts {
                    // eprintln!("Backend health check timeout");

                    let _ = child.kill();
                    let _ = child.wait();

                    return Err(std::io::Error::other(
                        "Backend health check timeout",
                    )
                    .into());
                }

                thread::sleep(Duration::from_secs(1));
            }

            // Keep the child handle managed by Tauri.
            let handle = BackendHandle {
                child: Mutex::new(Some(child)),
            };
            app.manage(handle);

            Ok(())
        })
        .on_window_event(|app, event| {
            use tauri::WindowEvent;
            if let WindowEvent::CloseRequested { api: _, .. } = event {
                if let Some(handle) = app.try_state::<BackendHandle>() {
                    if let Ok(mut guard) = handle.child.lock() {
                        if let Some(mut child) = guard.take() {
                            let _ = child.kill();
                            let _ = child.wait();
                        }
                    }
                }
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running Tauri application");
}

