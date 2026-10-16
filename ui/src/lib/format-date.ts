import { format, parseISO } from 'date-fns'

export function formatPanelDate(value?: string | null) {
  if (!value) return 'N/A'
  const trimmed = value.trim()
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) return trimmed
  try {
    return format(parseISO(trimmed), 'yyyy-MM-dd')
  } catch {
    return trimmed
  }
}

export function formatPanelDateTime(unix?: number | null) {
  if (!unix) return 'N/A'
  return format(new Date(unix * 1000), 'yyyy-MM-dd HH:mm')
}
