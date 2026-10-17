
use tauri::Manager;
use std::process::Command;
use std::thread;
use std::time::Duration;

#[cfg(windows)]
use std::os::windows::process::CommandExt;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            // Resolve Tauri resource directory.
            let resource_dir = app.path().resource_dir()?;

            // The installed MSI places the backend in:
            // C:\Program Files\FinanceApp\resources\financeapp-backend.exe
            let backend_path = resource_dir
                .join("resources")
                .join("financeapp-backend.exe");

            eprintln!("DEBUG Tauri resource_dir: {}", resource_dir.display());
            eprintln!("DEBUG Backend path: {}", backend_path.display());
            eprintln!(
                "DEBUG Current dir: {}",
                std::env::current_dir()
                    .unwrap_or_default()
                    .display()
            );

            // Fail early with a useful error if the executable is missing.
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

            // Run the backend from its own directory.
            let backend_dir = backend_path
                .parent()
                .expect("Backend path has no parent");

            let mut cmd = Command::new(&backend_path);
            cmd.current_dir(backend_dir);

            // Configure per-user application data paths.
            if let Ok(appdata) = std::env::var("APPDATA") {
                let base = std::path::Path::new(&appdata).join("FinanceApp");

                let db_path = base
                    .join("data")
                    .join("db")
                    .join("finance.db");

                let artifacts_path = base
                    .join("data")
                    .join("artifacts");

                let logs_path = base.join("logs");

                cmd.env("DATABASE_PATH", &db_path);
                cmd.env("ARTIFACTS_PATH", &artifacts_path);
                cmd.env("LOGS_PATH", &logs_path);
            }

            // Configure backend port and processor resources.
            cmd.env("PORT", "8081");

            let processor_script = resource_dir
                .join("processor")
                .join("src")
                .join("main.py");

            cmd.env("PROCESSOR_SCRIPT", &processor_script);

            let requirements_file = resource_dir
                .join("processor")
                .join("requirements-lock.txt");

            cmd.env("PROCESSOR_REQUIREMENTS", &requirements_file);

            // Optional bundled Python runtime archive.
            let python_bundle = resource_dir
                .join("python-runtime")
                .join("python-3.12.10-amd64.zip");

            if python_bundle.exists() {
                cmd.env("PYTHON_BUNDLE_PATH", &python_bundle);
            }

            // Optional offline Python wheels.
            let wheels_path = resource_dir.join("python-wheels");

            if wheels_path.exists() {
                cmd.env("PYTHON_WHEELS_PATH", &wheels_path);
            }

            // Prevent a separate console window on Windows.
            #[cfg(windows)]
            unsafe {
                cmd.creation_flags(0x08000000); // CREATE_NO_WINDOW
            }

            // Start backend.
            let mut child = cmd.spawn().map_err(|e| {
                eprintln!("Failed to start backend: {}", e);
                std::io::Error::other(format!(
                    "Failed to start backend at {}: {}",
                    backend_path.display(),
                    e
                ))
            })?;

            eprintln!("Backend started with PID {}", child.id());

            // Wait for backend health endpoint.
            let health_url = "http://127.0.0.1:8081/health";
            let mut attempts = 0;
            let max_attempts = 60;

            loop {
                attempts += 1;

                if let Ok(resp) = reqwest::blocking::get(health_url) {
                    if resp.status().is_success() {
                        eprintln!(
                            "Backend health OK after {} attempts",
                            attempts
                        );
                        break;
                    }
                }

                if attempts >= max_attempts {
                    eprintln!("Backend health check timeout");

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
            app.manage(child);

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
