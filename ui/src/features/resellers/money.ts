export function formatToman(rial: string): string {
  const value = BigInt(rial),
    negative = value < 0n,
    absolute = negative ? -value : value
  const whole = new Intl.NumberFormat('en-US').format(absolute / 10n),
    fraction = absolute % 10n
  return `${negative ? '-' : ''}${whole}${fraction ? `.${fraction}` : ''}`
}
export function tomanToRial(value: string): string {
  if (!/^\d+(?:\.\d)?$/.test(value.trim()))
    throw new Error('Enter toman with at most one decimal place.')
  const [whole, fraction = '0'] = value.trim().split('.'),
    rial = BigInt(whole) * 10n + BigInt(fraction)
  if (rial > 9223372036854775807n) throw new Error('Amount is too large.')
  return rial.toString()
}
export function gibToBytes(value: string): string | null {
  if (!value.trim()) return null
  if (!/^\d+(?:\.\d+)?$/.test(value))
    throw new Error('Enter a valid GiB quantity.')
  const [whole, fraction = ''] = value.split('.'),
    denominator = 10n ** BigInt(fraction.length),
    numerator =
      (BigInt(whole) * denominator + BigInt(fraction || '0')) * (1n << 30n)
  if (
    numerator % denominator !== 0n ||
    numerator / denominator > 9223372036854775807n
  )
    throw new Error('Quota must resolve to whole bytes within range.')
  return (numerator / denominator).toString()
}
export function bytesToGiB(value: string | null): string {
  if (value === null) return ''
  const bytes = BigInt(value),
    base = 1n << 30n,
    whole = bytes / base,
    remainder = bytes % base
  if (!remainder) return whole.toString()
  return `${whole}.${((remainder * 10n ** 30n) / base).toString().padStart(30, '0').replace(/0+$/, '')}`
}
export function dateLabel(epoch: number): string {
  return epoch ? new Date(epoch * 1000).toLocaleString() : '—'
}
export function inputToman(rial: string): string {
  const value = BigInt(rial),
    whole = value / 10n,
    fraction = value % 10n
  return `${whole}${fraction ? `.${fraction < 0n ? -fraction : fraction}` : ''}`
}
