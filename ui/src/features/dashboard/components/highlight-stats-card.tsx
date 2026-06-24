import { ReactNode } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

type HighlightStatsCardProps = {
  title: string
  icon: ReactNode
  value: number | string | undefined
  suffix?: string
  isLoading: boolean
  action?: ReactNode
}

export function HighlightStatsCard({
  title,
  icon,
  value,
  suffix,
  isLoading,
  action,
}: HighlightStatsCardProps) {
  let displayValue: string | number = '-'
  if (value !== undefined && value !== null && value !== '') {
    displayValue = suffix ? `${value} ${suffix}` : value
  }

  return (
    <Card className='border-border/60 py-0'>
      <CardContent className='flex items-center justify-between gap-5 px-6 py-5'>
        <div className='flex min-w-0 items-center gap-5'>
          <div className='bg-white/10 text-white flex size-11 shrink-0 items-center justify-center rounded-lg [&>svg]:size-5'>
            {icon}
          </div>
          <div className='min-w-0'>
            <p className='truncate text-lg font-medium leading-snug text-gray-300'>
              {title}
            </p>
            {isLoading ? (
              <Skeleton className='mt-1.5 h-7 w-24 rounded-sm' />
            ) : (
              <p className='truncate text-2xl font-bold leading-snug tracking-tight text-white'>
                {displayValue}
              </p>
            )}
          </div>
        </div>
        {action}
      </CardContent>
    </Card>
  )
}
