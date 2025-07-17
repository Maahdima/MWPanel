import { DeviceData } from '@/schema/dashboard.ts'
import { buildDeviceStats } from '@/utils/helper.ts'
import { Card, CardContent } from '@/components/ui/card'

interface DeviceResourceProps {
  stats: DeviceData | undefined
}

export default function DeviceResource({ stats }: DeviceResourceProps) {
  const items = buildDeviceStats(stats)

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
              <div className='flex flex-col items-end text-sm'>
                <span
                  className={
                    value === 'N/A' ? 'text-muted-foreground' : 'text-white'
                  }
                >
                  {value}
                </span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
