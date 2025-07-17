import { DeviceData } from '@/schema/dashboard.ts'

function getValueOrNA(value: string | number | undefined | null): string {
  if (value === undefined || value === null || value === '') return 'N/A'
  return String(value)
}

function formatGB(bytes: string | number | undefined): string | null {
  const num = Number(bytes)
  return isNaN(num) ? null : (num / 1_000_000_000).toFixed(2)
}

export function buildDeviceStats(
  stats: DeviceData | undefined
): { label: string; value: string }[] {
  const deviceInfo = stats?.DeviceInfo

  const uptime = getValueOrNA(deviceInfo?.uptime)
  const cpuLoad =
    deviceInfo?.cpu_load != null ? `${deviceInfo.cpu_load}%` : 'N/A'

  const memoryUsed = formatGB(
    Number(deviceInfo?.total_memory) - Number(deviceInfo?.free_memory)
  )
  const memoryTotal = formatGB(deviceInfo?.total_memory)
  const memoryValue =
    memoryUsed && memoryTotal ? `${memoryUsed}/${memoryTotal} GB` : 'N/A'

  const diskUsed = formatGB(
    Number(deviceInfo?.total_disk) - Number(deviceInfo?.free_disk)
  )
  const diskTotal = formatGB(deviceInfo?.total_disk)
  const diskValue =
    diskUsed && diskTotal ? `${diskUsed}/${diskTotal} GB` : 'N/A'

  return [
    { label: 'Uptime', value: uptime },
    { label: 'CPU Load', value: cpuLoad },
    { label: 'Memory Usage', value: memoryValue },
    { label: 'Disk Usage', value: diskValue },
  ]
}
