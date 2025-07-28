import * as React from 'react'
import { Area, AreaChart, CartesianGrid, XAxis } from 'recharts'
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

// TODO : implement this
const chartData = [
  { date: '2024-04-01', download: '222', upload: '150', total: '372' },
  { date: '2024-04-02', download: 97, upload: 180, total: 277 },
  { date: '2024-04-03', download: 167, upload: 120, total: 287 },
  { date: '2024-04-04', download: 242, upload: 260, total: 502 },
  { date: '2024-04-05', download: 373, upload: 290, total: 663 },
  { date: '2024-04-06', download: 301, upload: 340, total: 641 },
  { date: '2024-04-07', download: 245, upload: 180, total: 425 },
  { date: '2024-04-08', download: 409, upload: 320, total: 729 },
  { date: '2024-04-09', download: 59, upload: 110, total: 169 },
  { date: '2024-04-10', download: 261, upload: 190, total: 451 },
  { date: '2024-04-11', download: 327, upload: 350, total: 677 },
  { date: '2024-04-12', download: 292, upload: 210, total: 502 },
  { date: '2024-04-13', download: 342, upload: 380, total: 722 },
  { date: '2024-04-14', download: 137, upload: 220, total: 357 },
  { date: '2024-04-15', download: 120, upload: 170, total: 290 },
  { date: '2024-04-16', download: 138, upload: 190, total: 328 },
  { date: '2024-04-17', download: 446, upload: 360, total: 806 },
  { date: '2024-04-18', download: 364, upload: 410, total: 774 },
  { date: '2024-04-19', download: 243, upload: 180, total: 423 },
  { date: '2024-04-20', download: 89, upload: 150, total: 239 },
  { date: '2024-04-21', download: 137, upload: 200, total: 337 },
  { date: '2024-04-22', download: 224, upload: 170, total: 394 },
  { date: '2024-04-23', download: 138, upload: 230, total: 368 },
  { date: '2024-04-24', download: 387, upload: 290, total: 677 },
  { date: '2024-04-25', download: 215, upload: 250, total: 465 },
  { date: '2024-04-26', download: 75, upload: 130, total: 205 },
  { date: '2024-04-27', download: 383, upload: 420, total: 803 },
  { date: '2024-04-28', download: 122, upload: 180, total: 302 },
  { date: '2024-04-29', download: 315, upload: 240, total: 555 },
  { date: '2024-04-30', download: 454, upload: 380, total: 834 },
  { date: '2024-05-01', download: 165, upload: 220, total: 385 },
  { date: '2024-05-02', download: 293, upload: 310, total: 603 },
  { date: '2024-05-03', download: 247, upload: 190, total: 437 },
  { date: '2024-05-04', download: 385, upload: 420, total: 805 },
  { date: '2024-05-05', download: 481, upload: 390, total: 871 },
  { date: '2024-05-06', download: 498, upload: 520, total: 1018 },
  { date: '2024-05-07', download: 388, upload: 300, total: 688 },
  { date: '2024-05-08', download: 149, upload: 210, total: 359 },
  { date: '2024-05-09', download: 227, upload: 180, total: 407 },
  { date: '2024-05-10', download: 293, upload: 330, total: 623 },
  { date: '2024-05-11', download: 335, upload: 270, total: 605 },
  { date: '2024-05-12', download: 197, upload: 240, total: 437 },
  { date: '2024-05-13', download: 197, upload: 160, total: 357 },
  { date: '2024-05-14', download: 448, upload: 490, total: 938 },
  { date: '2024-05-15', download: 473, upload: 380, total: 853 },
  { date: '2024-05-16', download: 338, upload: 400, total: 738 },
  { date: '2024-05-17', download: 499, upload: 420, total: 919 },
  { date: '2024-05-18', download: 315, upload: 350, total: 665 },
  { date: '2024-05-19', download: 235, upload: 180, total: 415 },
  { date: '2024-05-20', download: 177, upload: 230, total: 407 },
  { date: '2024-05-21', download: 82, upload: 140, total: 222 },
  { date: '2024-05-22', download: 81, upload: 120, total: 201 },
  { date: '2024-05-23', download: 252, upload: 290, total: 542 },
  { date: '2024-05-24', download: 294, upload: 220, total: 514 },
  { date: '2024-05-25', download: 201, upload: 250, total: 451 },
  { date: '2024-05-26', download: 213, upload: 170, total: 383 },
  { date: '2024-05-27', download: 420, upload: 460, total: 880 },
  { date: '2024-05-28', download: 233, upload: 190, total: 423 },
  { date: '2024-05-29', download: 78, upload: 130, total: 208 },
  { date: '2024-05-30', download: 340, upload: 280, total: 620 },
  { date: '2024-05-31', download: 178, upload: 230, total: 408 },
  { date: '2024-06-01', download: 178, upload: 200, total: 378 },
  { date: '2024-06-02', download: 470, upload: 410, total: 880 },
  { date: '2024-06-03', download: 103, upload: 160, total: 263 },
  { date: '2024-06-04', download: 439, upload: 380, total: 819 },
  { date: '2024-06-05', download: 88, upload: 140, total: 228 },
  { date: '2024-06-06', download: 294, upload: 250, total: 544 },
  { date: '2024-06-07', download: 323, upload: 370, total: 693 },
  { date: '2024-06-08', download: 385, upload: 320, total: 705 },
  { date: '2024-06-09', download: 438, upload: 480, total: 918 },
  { date: '2024-06-10', download: 155, upload: 200, total: 355 },
  { date: '2024-06-11', download: 92, upload: 150, total: 242 },
  { date: '2024-06-12', download: 492, upload: 420, total: 912 },
  { date: '2024-06-13', download: 81, upload: 130, total: 211 },
  { date: '2024-06-14', download: 426, upload: 380, total: 806 },
  { date: '2024-06-15', download: 307, upload: 350, total: 657 },
  { date: '2024-06-16', download: 371, upload: 310, total: 681 },
  { date: '2024-06-17', download: 475, upload: 520, total: 995 },
  { date: '2024-06-18', download: 107, upload: 170, total: 277 },
  { date: '2024-06-19', download: 341, upload: 290, total: 631 },
  { date: '2024-06-20', download: 408, upload: 450, total: 858 },
  { date: '2024-06-21', download: 169, upload: 210, total: 379 },
  { date: '2024-06-22', download: 317, upload: 270, total: 587 },
  { date: '2024-06-23', download: 480, upload: 530, total: 1010 },
  { date: '2024-06-24', download: 132, upload: 180, total: 312 },
  { date: '2024-06-25', download: 141, upload: 190, total: 331 },
  { date: '2024-06-26', download: 434, upload: 380, total: 814 },
  { date: '2024-06-27', download: 448, upload: 490, total: 938 },
  { date: '2024-06-28', download: 149, upload: 200, total: 349 },
  { date: '2024-06-29', download: 103, upload: 160, total: 263 },
  { date: '2024-06-30', download: 446, upload: 400, total: 846 },
]

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
  const [timeRange, setTimeRange] = React.useState('90d')

  const filteredData = chartData.filter((item) => {
    const date = new Date(item.date)
    const referenceDate = new Date('2024-06-30')
    let daysToSubtract = 90
    if (timeRange === '30d') {
      daysToSubtract = 30
    } else if (timeRange === '7d') {
      daysToSubtract = 7
    }
    const startDate = new Date(referenceDate)
    startDate.setDate(startDate.getDate() - daysToSubtract)
    return date >= startDate
  })

  return (
    <Card className='col-span-1 pt-0 lg:col-span-5'>
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
      <CardContent className='px-2 pt-4 sm:px-6 sm:pt-6'>
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
                  labelFormatter={(value) => {
                    return new Date(value).toLocaleDateString('en-US', {
                      month: 'short',
                      day: 'numeric',
                    })
                  }}
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
            <ChartLegend content={<ChartLegendContent payload={undefined} />} />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
