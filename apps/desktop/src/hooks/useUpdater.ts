import { useCallback, useState } from 'react';

import { check, type DownloadEvent, type Update } from '@tauri-apps/plugin-updater';
import { progressLabel } from '../lib/updateProgress';

/* The updater state machine: check, offer, download, install.

   The plugin emits Started (with the total bytes) and one Progress
   event per chunk — there is no percentage in the event, so the label
   accumulates the chunks (progressLabel), and without a known total the
   honest label is the downloaded size. Settings composes the UI; the
   machine lives here. */

export interface Updater {
  checking: boolean;
  available: Update | null;
  confirming: boolean;
  progress: string | null;
  error: string | null;
  checkUpdates: () => Promise<void>;
  installUpdate: () => Promise<void>;
  cancelInstall: () => void;
}

export function useUpdater(): Updater {
  const [checking, setChecking] = useState(false);
  const [available, setAvailable] = useState<Update | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [progress, setProgress] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const checkUpdates = useCallback(async () => {
    setChecking(true);
    setError(null);
    setProgress(null);
    try {
      const update = await check();
      if (update) {
        setAvailable(update);
        setConfirming(true);
      } else {
        setProgress('FinanceApp está actualizado.');
        setTimeout(() => setProgress(null), 3000);
      }
    } catch {
      setError('Error verificando actualizaciones.');
    } finally {
      setChecking(false);
    }
  }, []);

  const installUpdate = useCallback(async () => {
    if (!available) return;
    setConfirming(false);
    setProgress(`Descargando versión ${available.version}...`);
    try {
      // El plugin permite callback de progreso (Started trae el total;
      // Progress trae cada chunk: se acumula, no existe un percent).
      let downloaded = 0;
      let total = 0;
      await available.downloadAndInstall((event: DownloadEvent) => {
        if (event.event === 'Started') {
          total = event.data.contentLength ?? 0;
        }
        if (event.event === 'Progress') {
          downloaded += event.data.chunkLength;
          setProgress(`Descargando ${available.version}... ${progressLabel(downloaded, total)}`);
        }
      });
      setProgress('Instalación completada. Reiniciando...');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Falló la instalación de la actualización.');
      setProgress(null);
    }
  }, [available]);

  const cancelInstall = useCallback(() => setConfirming(false), []);

  return {
    checking,
    available,
    confirming,
    progress,
    error,
    checkUpdates,
    installUpdate,
    cancelInstall,
  };
}
