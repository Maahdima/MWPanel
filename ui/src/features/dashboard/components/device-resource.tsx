import { DeviceData } from '@/schema/dashboard.ts'
import { Card, CardContent } from '@/components/ui/card'

interface DeviceResourceProps {
  stats: DeviceData | null
}

export default function DeviceResource({ stats }: DeviceResourceProps) {
  const items = [
    { label: 'Uptime', value: stats?.DeviceInfo.uptime },
    {
      label: 'CPU Load',
      value: `${stats?.DeviceInfo.cpu_load}%`,
    },
    {
      label: 'Memory Usage',
      value: `${stats ? ((Number(stats.DeviceInfo.total_memory) - Number(stats.DeviceInfo.free_memory)) / 1_000_000_000).toFixed(2) : 'N/A'}/${stats ? (Number(stats.DeviceInfo.total_memory) / 1_000_000_000).toFixed(2) : 'N/A'} GB`,
    },
    {
      label: 'Disk Usage',
      value: stats
        ? `${((Number(stats.DeviceInfo.total_disk) - Number(stats.DeviceInfo.free_disk)) / 1_000_000_000).toFixed(2)}/${(Number(stats.DeviceInfo.total_disk) / 1_000_000_000).toFixed(2)} GB`
        : 'N/A',
    },
  ]

  return (
    <Card className='col-span-1 lg:col-span-2'>
      <CardContent>
        <h2 className='mb-4 text-lg font-semibold'>Hardware Statistics</h2>
        <div>
          {items.map(({ label, value }, idx) => (
            <div
              key={idx}
              className='flex items-start justify-between border-b border-white/10 py-4'
            >
              <span className='text-sm font-medium text-gray-300'>{label}</span>
              <div className='flex flex-col items-end text-sm text-white'>
                <span>{value}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
