use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let handle = app.handle();
            std::thread::spawn(move || {
                let _ = std::process::Command::new("financeapp-backend.exe")
                    .current_dir("resources")
                    .spawn();
            });
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
