use tauri::Manager;
use std::process::Command;
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
            }

            #[cfg(windows)]
            unsafe { cmd.creation_flags(0x08000000); } // CREATE_NO_WINDOW

            match cmd.spawn() {
                Ok(child) => {
                    eprintln!("Backend started with PID {}", child.id());
                }
                Err(e) => {
                    eprintln!("Failed to start backend: {}", e);
                    return Err(Box::new(std::io::Error::other(e.to_string())));
                }
            }

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
