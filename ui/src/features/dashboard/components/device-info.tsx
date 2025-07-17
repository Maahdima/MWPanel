import { DeviceData } from '@/schema/dashboard.ts'
import { Card, CardContent } from '@/components/ui/card'

interface DeviceInfoProps {
  stats: DeviceData | undefined
}

export default function DeviceInfo({ stats }: DeviceInfoProps) {
  const items = [
    { label: 'Identity', value: stats?.DeviceIdentity.identity },
    {
      label: 'Device',
      value: stats?.DeviceInfo.board_name,
      badge: stats?.DeviceInfo.cpu_arch,
    },
    {
      label: 'OS Version',
      value: stats?.DeviceInfo.os_version,
      badge: 'stable', //  make dynamic if needed
      badgeStyle: 'bg-yellow-500',
    },
    {
      label: 'Public IPv4',
      value: stats?.DeviceIPv4Address?.ipv4,
      badge: stats?.DeviceIPv4Address?.isp,
      badgeStyle: 'bg-cyan-700',
    },
    { label: 'DNS Servers', value: stats?.DNSConfig?.dns_servers },
  ]

  return (
    <Card className='col-span-1 lg:col-span-3'>
      <CardContent>
        <h2 className='mb-4 text-lg font-semibold'>Mikrotik Statistics</h2>
        <div>
          {items.map(({ label, value, badge, badgeStyle }, idx) => (
            <div
              key={idx}
              className='flex items-start justify-between border-b border-white/10 py-4'
            >
              <span className='text-sm font-medium text-gray-300'>{label}</span>
              <div className='flex flex-col items-end text-sm text-white'>
                <span className={value ?? 'text-muted-foreground'}>
                  {value ?? 'N/A'}
                </span>
                {badge && (
                  <span
                    className={`mt-1 rounded-md px-2 py-0.5 text-xs text-white ${
                      badgeStyle || 'bg-purple-600'
                    }`}
                  >
                    {badge}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
