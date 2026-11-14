import prettier from '@vue/eslint-config-prettier'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginVue from 'eslint-plugin-vue'

const config: ReturnType<typeof defineConfigWithVueTs> = defineConfigWithVueTs(
  {
    name: 'simon-ostendorf-de/files',
    files: ['**/*.ts', '**/*.vue'],
  },
  {
    name: 'simon-ostendorf-de/ignores',
    ignores: ['dist/**', 'node_modules/**'],
  },
  pluginVue.configs['flat/recommended'],
  vueTsConfigs.recommendedTypeChecked,
  prettier,
  {
    name: 'simon-ostendorf-de/rules',
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-non-null-assertion': 'error',
      '@typescript-eslint/consistent-type-imports': 'error',
    },
  },
  {
    name: 'simon-ostendorf-de/tests',
    files: ['**/*.test.ts'],
    rules: {
      '@typescript-eslint/unbound-method': 'off',
      'vue/one-component-per-file': 'off',
    },
  },
)

export default config
