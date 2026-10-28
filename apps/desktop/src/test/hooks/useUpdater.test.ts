import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { check, type DownloadEvent, type Update } from '@tauri-apps/plugin-updater';
import { useUpdater } from '../../hooks/useUpdater';

/* The updater state machine: check, offer, download, install. The plugin
   events (Started with the total, Progress per chunk) arrive through the
   download callback; the machine turns them into the label Settings
   shows. */

vi.mock('@tauri-apps/plugin-updater', () => ({ check: vi.fn() }));

function updateMock(): Update {
  const downloadAndInstall = vi.fn(async (onEvent: (event: DownloadEvent) => void) => {
    onEvent({ event: 'Started', data: { contentLength: 1000 } } as DownloadEvent);
    onEvent({ event: 'Progress', data: { chunkLength: 500 } } as DownloadEvent);
    onEvent({ event: 'Progress', data: { chunkLength: 500 } } as DownloadEvent);
  });
  return { version: '0.2.0', downloadAndInstall } as unknown as Update;
}

describe('useUpdater', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('offers an available update for confirmation', async () => {
    vi.mocked(check).mockResolvedValue(updateMock());

    const { result } = renderHook(() => useUpdater());

    await act(async () => {
      await result.current.checkUpdates();
    });

    expect(result.current.available?.version).toBe('0.2.0');
    expect(result.current.confirming).toBe(true);
    expect(result.current.checking).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('reports "up to date" when the plugin finds nothing, and clears the message after 3s', async () => {
    vi.mocked(check).mockResolvedValue(null);

    const { result } = renderHook(() => useUpdater());

    await act(async () => {
      await result.current.checkUpdates();
    });
    expect(result.current.progress).toBe('FinanceApp está actualizado.');

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(result.current.progress).toBeNull();
  });

  it('reports a failed check as an error, not a silent state', async () => {
    vi.mocked(check).mockRejectedValue(new Error('network'));

    const { result } = renderHook(() => useUpdater());

    await act(async () => {
      await result.current.checkUpdates();
    });

    expect(result.current.error).toBe('Error verificando actualizaciones.');
    expect(result.current.available).toBeNull();
  });

  it('downloads with progress from the plugin events and reports completion', async () => {
    vi.mocked(check).mockResolvedValue(updateMock());

    const { result } = renderHook(() => useUpdater());
    await act(async () => {
      await result.current.checkUpdates();
    });

    await act(async () => {
      await result.current.installUpdate();
    });

    expect(result.current.available?.downloadAndInstall).toHaveBeenCalledTimes(1);
    expect(result.current.progress).toBe('Instalación completada. Reiniciando...');
    expect(result.current.error).toBeNull();
  });

  it('shows the accumulated download percentage while the chunks arrive', async () => {
    let emit: ((event: DownloadEvent) => void) | null = null;
    let finish: (() => void) | null = null;
    const downloadAndInstall = vi.fn(
      (onEvent: (event: DownloadEvent) => void) =>
        new Promise<void>((resolve) => {
          emit = onEvent;
          finish = resolve;
        }),
    );
    vi.mocked(check).mockResolvedValue({
      version: '0.2.0',
      downloadAndInstall,
    } as unknown as Update);

    const { result } = renderHook(() => useUpdater());
    await act(async () => {
      await result.current.checkUpdates();
    });

    // The install is in flight and parked: the callback is ours to drive.
    await act(async () => {
      void result.current.installUpdate();
    });
    expect(result.current.progress).toContain('Descargando versión 0.2.0');

    act(() => {
      emit!({ event: 'Started', data: { contentLength: 1000 } } as DownloadEvent);
      emit!({ event: 'Progress', data: { chunkLength: 250 } } as DownloadEvent);
    });

    expect(result.current.progress).toContain('25%');

    await act(async () => {
      finish!();
    });
    expect(result.current.progress).toBe('Instalación completada. Reiniciando...');
  });

  it('does nothing when there is no update to install', async () => {
    const { result } = renderHook(() => useUpdater());

    await act(async () => {
      await result.current.installUpdate();
    });

    expect(result.current.progress).toBeNull();
  });

  it('surfaces the plugin failure message and clears the progress', async () => {
    const update = updateMock();
    vi.mocked(update.downloadAndInstall).mockRejectedValue(new Error('msiexec failed'));
    vi.mocked(check).mockResolvedValue(update);

    const { result } = renderHook(() => useUpdater());
    await act(async () => {
      await result.current.checkUpdates();
    });

    await act(async () => {
      await result.current.installUpdate();
    });

    expect(result.current.error).toBe('msiexec failed');
    expect(result.current.progress).toBeNull();
  });

  it('cancels the offer without installing', async () => {
    vi.mocked(check).mockResolvedValue(updateMock());

    const { result } = renderHook(() => useUpdater());
    await act(async () => {
      await result.current.checkUpdates();
    });

    act(() => result.current.cancelInstall());

    expect(result.current.confirming).toBe(false);
    expect(result.current.available).not.toBeNull();
  });
});
