'use client'

import { DeviceData } from '@/schema/dashboard.ts'
import { Pie, PieChart } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'

type PeerStatus = 'online' | 'offline' | 'disabled'

const chartConfig: Record<PeerStatus, { label: string; color: string }> = {
  online: {
    label: 'Online',
    color: '#22c55e', // green-500
  },
  offline: {
    label: 'Offline',
    color: '#ef4444', // red-500
  },
  disabled: {
    label: 'Disabled',
    color: '#facc15', // yellow-400
  },
}

interface Props {
  stats: DeviceData['PeerInfo'] | undefined
}

export function PeersChart({ stats }: Props) {
  const chartData: { status: PeerStatus; count: number; fill: string }[] = [
    {
      status: 'online',
      count: stats?.online_peers ?? 0,
      fill: chartConfig.online.color,
    },
    {
      status: 'offline',
      count: stats?.offline_peers ?? 0,
      fill: chartConfig.offline.color,
    },
    {
      status: 'disabled',
      count: stats?.disabled_peers ?? 0,
      fill: chartConfig.disabled.color,
    },
  ]

  return (
    <Card className='col-span-1 flex flex-col lg:col-span-2'>
      <CardHeader className='items-center pb-0'>
        <CardTitle>
          <h2 className='mb-4 text-lg font-semibold'>Users Statistics</h2>
        </CardTitle>
      </CardHeader>
      <CardContent className='flex-1 pb-0'>
        <ChartContainer
          config={chartConfig}
          className='mx-auto aspect-square max-h-[250px]'
        >
          <PieChart>
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Pie
              data={chartData}
              dataKey='count'
              nameKey='status'
              innerRadius={60}
              strokeWidth={5}
            />
          </PieChart>
        </ChartContainer>

        <div className='mt-6 flex flex-wrap justify-center gap-4'>
          {chartData.map((item) => {
            const config = chartConfig[item.status]
            return (
              <div key={item.status} className='flex items-center gap-2'>
                <span
                  className='inline-block h-3 w-3 rounded-full'
                  style={{ backgroundColor: config.color }}
                />
                <span className='text-muted-foreground text-sm'>
                  {config.label}: {item.count.toLocaleString()}
                </span>
              </div>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
