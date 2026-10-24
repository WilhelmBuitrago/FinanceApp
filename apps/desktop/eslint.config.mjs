import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';

export default tseslint.config(
  { ignores: ['dist', 'node_modules', 'src-tauri'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: { 'react-hooks': reactHooks, 'react-refresh': reactRefresh },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      // Review-queue payloads are free-form backend JSON by design; typing
      // them is tracked debt, so this warns instead of blocking the lint.
      '@typescript-eslint/no-explicit-any': 'warn',
      // The React Compiler-era rules are on but phased: fixing the five
      // setState-in-effect sites (the fetch hooks) and the Date.now() memo
      // means refactors with behavior risk, tracked as debt instead of
      // being rushed in a tooling pass.
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/purity': 'warn',
    },
  },
);
