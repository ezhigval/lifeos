import { renderApp } from '@/root'
import { hydrateDesktopSession, markDesktopApp } from '@/lib/desktopSession'
import { DesktopFrame } from './DesktopFrame'
import './app.css'

// The window origin is stable, and the login is also written to
// Application Support so a webview wipe does not ask for the code again.
markDesktopApp()
void hydrateDesktopSession().finally(() => {
  renderApp('/', DesktopFrame)
})
