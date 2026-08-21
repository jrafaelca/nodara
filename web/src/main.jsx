import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.jsx'
import { ThemeProvider } from './components/theme-provider.jsx'
import { AuthProvider } from './lib/auth.jsx'
import './styles.css'

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ThemeProvider defaultTheme="light" storageKey="nodara-ui-theme">
      <AuthProvider><App /></AuthProvider>
    </ThemeProvider>
  </React.StrictMode>,
)
