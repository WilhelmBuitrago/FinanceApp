import { check } from '@tauri-apps/plugin-updater';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { purgeAllData } from '../../api/admin';
import {
  createCategory,
  deleteCategory as deleteCategoryApi,
  getCategories,
} from '../../api/categories';
import SettingsPage from '../../pages/SettingsPage';

vi.mock('../../api/admin', () => ({ purgeAllData: vi.fn() }));
vi.mock('../../api/categories', () => ({
  getCategories: vi.fn(),
  createCategory: vi.fn(),
  deleteCategory: vi.fn(),
}));
vi.mock('@tauri-apps/plugin-updater', () => ({ check: vi.fn() }));

const incomeCategory = { ID: 'cat_in', Name: 'Salario', CategoryType: 'INCOME' };
const expenseCategory = { ID: 'cat_ex', Name: 'Comida', CategoryType: 'EXPENSE' };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getCategories).mockResolvedValue([incomeCategory, expenseCategory] as never);
  vi.mocked(createCategory).mockResolvedValue({
    ID: 'cat_new',
    Name: 'Transporte',
    CategoryType: 'EXPENSE',
  } as never);
  vi.mocked(deleteCategoryApi).mockResolvedValue(undefined);
  vi.mocked(purgeAllData).mockResolvedValue(undefined);
});

describe('SettingsPage', () => {
  it('renders the backend catalog split by type', async () => {
    render(<SettingsPage />);
    await waitFor(() => {
      expect(screen.getByText('Salario')).toBeTruthy();
      expect(screen.getByText('Comida')).toBeTruthy();
    });
    expect(screen.getByRole('heading', { name: 'Categorías de Ingresos' })).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Categorías de Egresos' })).toBeTruthy();
  });

  it('creates a category through the api and shows it', async () => {
    render(<SettingsPage />);
    await waitFor(() => expect(screen.getByText('Comida')).toBeTruthy());

    fireEvent.change(screen.getByPlaceholderText('Nombre categoría'), {
      target: { value: 'Transporte' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Agregar' }));

    await waitFor(() => expect(screen.getByText('Transporte')).toBeTruthy());
    expect(createCategory).toHaveBeenCalledTimes(1);
    expect(vi.mocked(createCategory).mock.calls[0][1]).toBe('Transporte');
  });

  it('deletes a category through the api', async () => {
    render(<SettingsPage />);
    await waitFor(() => expect(screen.getByText('Comida')).toBeTruthy());

    // Scope to the expense column: both columns render an "Eliminar".
    const expenseColumn = screen.getByRole('heading', {
      name: 'Categorías de Egresos',
    }).parentElement!;
    fireEvent.click(within(expenseColumn).getByRole('button', { name: 'Eliminar' }));

    await waitFor(() => expect(deleteCategoryApi).toHaveBeenCalledWith('cat_ex'));
    await waitFor(() => expect(screen.queryByText('Comida')).toBeNull());
  });

  it('purges only after the destructive confirmation', async () => {
    render(<SettingsPage />);
    await waitFor(() => expect(screen.getByText('Comida')).toBeTruthy());

    // No modal before the click, no call without it.
    expect(screen.queryByRole('alertdialog')).toBeNull();
    expect(purgeAllData).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Borrar todos los datos' }));
    const modal = screen.getByRole('alertdialog');
    expect(within(modal).getByText(/no se puede\s+deshacer/i)).toBeTruthy();

    fireEvent.click(within(modal).getByRole('button', { name: 'Borrar todos los datos' }));

    await waitFor(() => expect(purgeAllData).toHaveBeenCalledTimes(1));
    expect(await screen.findByText('Todos los datos fueron eliminados.')).toBeTruthy();
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });

  it('reports "up to date" when the updater finds nothing', async () => {
    vi.mocked(check).mockResolvedValue(null);
    render(<SettingsPage />);

    fireEvent.click(screen.getByRole('button', { name: 'Buscar Actualizaciones' }));

    expect(await screen.findByText('FinanceApp está actualizado.')).toBeTruthy();
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });

  it('offers the found update and installs it on confirm', async () => {
    const downloadAndInstall = vi.fn().mockResolvedValue(undefined);
    vi.mocked(check).mockResolvedValue({
      version: '0.2.0',
      downloadAndInstall,
    } as never);
    render(<SettingsPage />);

    fireEvent.click(screen.getByRole('button', { name: 'Buscar Actualizaciones' }));

    const modal = await screen.findByRole('alertdialog');
    expect(within(modal).getByText(/0\.2\.0/)).toBeTruthy();

    fireEvent.click(within(modal).getByRole('button', { name: 'Instalar' }));

    await waitFor(() => expect(downloadAndInstall).toHaveBeenCalledTimes(1));
    expect(await screen.findByText('Instalación completada. Reiniciando...')).toBeTruthy();
  });
});
