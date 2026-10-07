import { createContext, useContext, type ReactNode } from 'react'

export type DesktopFrame = (props: { children: ReactNode }) => ReactNode

const DesktopFrameContext = createContext<DesktopFrame | null>(null)

export function DesktopFrameProvider({
  frame,
  children,
}: {
  frame: DesktopFrame | null
  children: ReactNode
}) {
  return <DesktopFrameContext.Provider value={frame}>{children}</DesktopFrameContext.Provider>
}

export function useDesktopFrame() {
  return useContext(DesktopFrameContext)
}
