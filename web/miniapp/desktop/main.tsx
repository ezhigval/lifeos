import { renderApp } from '@/root'
import { DesktopFrame } from './DesktopFrame'
import './app.css'

// Screens stay in this window. /api is proxied to the server in dev,
// or sent to VITE_API_BASE when the desktop build cannot share an origin.
renderApp('/', DesktopFrame)
