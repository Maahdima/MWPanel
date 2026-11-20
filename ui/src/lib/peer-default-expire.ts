import { addDays, format } from 'date-fns'

/** Maps panel default expire days to an expire_time form value, or null for no expiry. */
export function defaultExpireTimeFromDays(
  defaultExpireDays: number | null | undefined
): string | null {
  if (defaultExpireDays == null) {
    return null
  }

  return format(addDays(new Date(), defaultExpireDays), 'yyyy-MM-dd')
}
