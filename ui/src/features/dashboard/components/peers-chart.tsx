'use client'

import { Pie, PieChart } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'

const chartData = [
  { status: 'online', count: 15, fill: 'var(--color-online)' },
  { status: 'offline', count: 40, fill: 'var(--color-offline)' },
  { status: 'disabled', count: 5, fill: 'var(--color-disabled)' },
]

const chartConfig = {
  peers: {
    label: 'Peers',
  },
  online: {
    label: 'Online',
    color: 'var(--chart-2)',
  },
  offline: {
    label: 'Offline',
    color: 'var(--chart-5)',
  },
  disabled: {
    label: 'Disabled',
    color: 'var(--chart-3)',
  },
} satisfies ChartConfig

export function PeersChart() {
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
