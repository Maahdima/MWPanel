import * as React from 'react'
import { CloudOff } from 'lucide-react'
import { Area, AreaChart, CartesianGrid, XAxis } from 'recharts'
import { useDailyTrafficUsageQuery } from '@/hooks/dashboard/useDailyTrafficUsageQuery.ts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const chartConfig = {
  peers: {
    label: 'Peers',
  },
  download: {
    label: 'Download',
    color: 'var(--chart-1)',
  },
  upload: {
    label: 'Upload',
    color: 'var(--chart-2)',
  },
  total: {
    label: 'Total',
    color: 'var(--chart-3)',
  },
} satisfies ChartConfig

export function ChartAreaInteractive() {
  const [timeRange, setTimeRange] = React.useState('7d')

  const rangeNumber = React.useMemo(
    () => parseInt(timeRange.replace('d', '')),
    [timeRange]
  )

  const {
    data: dailyTrafficUsageData,
    isLoading,
    isFetching,
  } = useDailyTrafficUsageQuery(rangeNumber)

  const filteredData = dailyTrafficUsageData ?? []

  return (
    <Card className='col-span-1 flex flex-col pt-0 lg:col-span-5'>
      <CardHeader className='flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row'>
        <div className='grid flex-1 gap-1'>
          <CardTitle>
            <h2 className='mb-4 text-lg font-semibold'>Daily Traffic Usage</h2>
          </CardTitle>
        </div>
        <Select value={timeRange} onValueChange={setTimeRange}>
          <SelectTrigger
            className='hidden w-[160px] rounded-lg sm:ml-auto sm:flex'
            aria-label='Select a value'
          >
            <SelectValue placeholder='Last 3 months' />
          </SelectTrigger>
          <SelectContent className='rounded-xl'>
            <SelectItem value='90d' className='rounded-lg'>
              Last 3 months
            </SelectItem>
            <SelectItem value='30d' className='rounded-lg'>
              Last 30 days
            </SelectItem>
            <SelectItem value='7d' className='rounded-lg'>
              Last 7 days
            </SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>

      <CardContent className='flex flex-1 px-2 pt-4 sm:px-6 sm:pt-6'>
        {isLoading || isFetching ? (
          <div className='flex h-full w-full animate-pulse flex-col'>
            <div className='flex-1'>
              <svg
                className='h-full w-full'
                viewBox='0 0 300 100'
                preserveAspectRatio='none'
              >
                <defs>
                  <linearGradient
                    id='skeleton-gradient'
                    x1='0'
                    y1='0'
                    x2='0'
                    y2='1'
                  >
                    <stop
                      offset='5%'
                      stopColor='hsl(var(--muted))'
                      stopOpacity={0.5}
                    />
                    <stop
                      offset='95%'
                      stopColor='hsl(var(--muted))'
                      stopOpacity={0}
                    />
                  </linearGradient>
                </defs>

                <path
                  d='M 0 60 Q 50 20 100 50 T 200 80 T 300 60 L 300 100 L 0 100 Z'
                  fill='url(#skeleton-gradient)'
                  stroke='hsl(var(--muted-foreground) / 0.4)'
                  strokeWidth='2'
                />
              </svg>
            </div>

            <div className='bg-muted-foreground/10 mt-2 h-px w-full' />
            <div className='flex items-center justify-between pt-2'>
              {[...Array(5)].map((_, i) => (
                <div key={i} className='bg-muted h-2 w-10 rounded-full' />
              ))}
            </div>
          </div>
        ) : filteredData.length === 0 ? (
          <div className='text-muted-foreground flex h-full w-full flex-col items-center justify-center'>
            <CloudOff className='mb-4 text-gray-300' size={64} />
            <p className='text-center text-base font-medium'>
              No traffic data available for the selected range.
            </p>
          </div>
        ) : (
          <ChartContainer
            config={chartConfig}
            className='aspect-auto h-[250px] w-full'
          >
            <AreaChart data={filteredData}>
              <defs>
                <linearGradient id='fillDownload' x1='0' y1='0' x2='0' y2='1'>
                  <stop
                    offset='5%'
                    stopColor='var(--color-download)'
                    stopOpacity={0.8}
                  />
                  <stop
                    offset='95%'
                    stopColor='var(--color-download)'
                    stopOpacity={0.1}
                  />
                </linearGradient>
                <linearGradient id='fillUpload' x1='0' y1='0' x2='0' y2='1'>
                  <stop
                    offset='5%'
                    stopColor='var(--color-upload)'
                    stopOpacity={0.8}
                  />
                  <stop
                    offset='95%'
                    stopColor='var(--color-upload)'
                    stopOpacity={0.1}
                  />
                </linearGradient>
                <linearGradient id='fillTotal' x1='0' y1='0' x2='0' y2='1'>
                  <stop
                    offset='5%'
                    stopColor='var(--color-total)'
                    stopOpacity={0.8}
                  />
                  <stop
                    offset='95%'
                    stopColor='var(--color-total)'
                    stopOpacity={0.1}
                  />
                </linearGradient>
              </defs>

              <CartesianGrid vertical={false} />
              <XAxis
                dataKey='date'
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={32}
                tickFormatter={(value) => {
                  const date = new Date(value)
                  return date.toLocaleDateString('en-US', {
                    month: 'short',
                    day: 'numeric',
                  })
                }}
              />

              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    labelFormatter={(value) =>
                      new Date(value).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                      })
                    }
                    indicator='dot'
                  />
                }
              />

              <Area
                dataKey='upload'
                type='natural'
                fill='url(#fillUpload)'
                stroke='var(--color-upload)'
                stackId='a'
              />
              <Area
                dataKey='download'
                type='natural'
                fill='url(#fillDownload)'
                stroke='var(--color-download)'
                stackId='a'
              />
              <Area
                dataKey='total'
                type='natural'
                fill='url(#fillTotal)'
                stroke='var(--color-total)'
                stackId='a'
              />

              <ChartLegend content={<ChartLegendContent />} />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
