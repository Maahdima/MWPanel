export type MwpRuntimeConfig = {
  adminPanelPath: string
}

declare global {
  interface Window {
    __MWP_CONFIG__?: MwpRuntimeConfig
  }
}

function normalizePath(path: string | undefined | null): string {
  if (!path || path === '/') return ''
  const trimmed = `/${String(path).replace(/^\/+|\/+$/g, '')}`
  return trimmed === '/' ? '' : trimmed
}

/** Resolves admin panel path from server injection or Vite env (dev). */
export function loadAppConfig(): MwpRuntimeConfig {
  const injected = window.__MWP_CONFIG__
  if (injected && typeof injected.adminPanelPath === 'string') {
    return { adminPanelPath: normalizePath(injected.adminPanelPath) }
  }

  const fromVite = import.meta.env.VITE_ADMIN_PANEL_PATH as string | undefined
  return { adminPanelPath: normalizePath(fromVite) }
}

export function adminLoginPath(adminPanelPath: string): string {
  const base = normalizePath(adminPanelPath)
  return base ? `${base}/sign-in` : '/sign-in'
}
