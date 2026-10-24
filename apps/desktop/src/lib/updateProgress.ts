/* Updater download progress. The plugin sends Started (with the total
   bytes) and one Progress event per chunk: there is no percentage in the
   event, so the display accumulates the chunks. Without a known total
   the honest label is the downloaded size. */

export function progressLabel(downloaded: number, total: number): string {
  if (total > 0) return `${Math.round((downloaded / total) * 100)}%`;
  return `${Math.round(downloaded / 1024)} KB`;
}
