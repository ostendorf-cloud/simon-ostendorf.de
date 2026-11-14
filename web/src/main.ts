import { createApp } from 'vue'
import { useTheme } from '@ostendorf-cloud/govue'

import App from './App.vue'
import './styles.css'

// The theme never varies: this site ships exactly one brand. Mode starts
// at "system", the visitor's own choice only overrides it once they use
// the toggle, matching useTheme's own documented default.
useTheme({ theme: 'simon-ostendorf-de' })

createApp(App).mount('#app')
