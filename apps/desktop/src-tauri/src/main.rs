use tauri::Manager;
use std::process::Command;
use std::time::Duration;
use std::thread;
#[cfg(windows)]
use std::os::windows::process::CommandExt;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            // Resolve backend executable path via Tauri resources
            let resource_dir = app.path().resource_dir()?;
            let backend_path = resource_dir.join("financeapp-backend.exe");

            // Build environment for backend with explicit data paths
            let mut cmd = Command::new(&backend_path);
            if let Ok(appdata) = std::env::var("APPDATA") {
                let base = std::path::Path::new(&appdata).join("FinanceApp");
                let db_path = base.join("data").join("db").join("finance.db");
                let artifacts_path = base.join("data").join("artifacts");
                let logs_path = base.join("logs");
                cmd.env("DATABASE_PATH", &db_path);
                cmd.env("ARTIFACTS_PATH", &artifacts_path);
                cmd.env("LOGS_PATH", &logs_path);
                cmd.env("PORT", "8081");
                // Processor script location
                let processor_script = resource_dir.join("processor").join("src").join("main.py");
                cmd.env("PROCESSOR_SCRIPT", &processor_script);
                let requirements_file = resource_dir.join("processor").join("requirements-lock.txt");
                cmd.env("PROCESSOR_REQUIREMENTS", &requirements_file);
                // Optional bundled python runtime
                let python_bundle = resource_dir.join("python-runtime").join("python-3.12.10-amd64.zip");
                if python_bundle.exists() {
                    cmd.env("PYTHON_BUNDLE_PATH", &python_bundle);
                }
                let wheels_path = resource_dir.join("python-wheels");
                if wheels_path.exists() {
                    cmd.env("PYTHON_WHEELS_PATH", &wheels_path);
                }
            }

            #[cfg(windows)]
            unsafe { cmd.creation_flags(0x08000000); } // CREATE_NO_WINDOW

            let mut child = cmd.spawn().map_err(|e| {
                eprintln!("Failed to start backend: {}", e);
                Box::new(std::io::Error::other(e.to_string())) as Box<dyn std::error::Error>
            })?;
            eprintln!("Backend started with PID {}", child.id());

            // Wait for backend health
            let health_url = "http://localhost:8081/health";
            let mut attempts = 0;
            let max_attempts = 60;
            loop {
                attempts += 1;
                if let Ok(resp) = reqwest::blocking::get(health_url) {
                    if resp.status().is_success() {
                        eprintln!("Backend health OK after {} attempts", attempts);
                        break;
                    }
                }
                if attempts >= max_attempts {
                    eprintln!("Backend health check timeout");
                    let _ = child.kill();
                    return Err(Box::new(std::io::Error::other("Backend health check timeout")));
                }
                thread::sleep(Duration::from_secs(1));
            }

            // Store child handle for cleanup on exit
            app.manage(child);

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
