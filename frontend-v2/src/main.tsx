import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'
import { PeriodFilterProvider } from './lib/periodFilter'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <PeriodFilterProvider>
      <App />
    </PeriodFilterProvider>
  </StrictMode>,
)
