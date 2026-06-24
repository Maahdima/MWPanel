import { JSX } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

type StatsCardProps = {
  title: string
  icon: JSX.Element
  value: number | string | undefined
  isLoading: boolean
}

export function StatsCard({ title, icon, value, isLoading }: StatsCardProps) {
  return (
    <Card className='border-border/60 py-0'>
      <CardContent className='flex items-center gap-5 px-6 py-5'>
        <div className='bg-white/10 text-white flex size-11 shrink-0 items-center justify-center rounded-lg [&>svg]:size-5'>
          {icon}
        </div>
        <div className='min-w-0'>
          <p className='truncate text-lg font-medium leading-snug text-gray-300'>
            {title}
          </p>
          {isLoading ? (
            <Skeleton className='mt-1.5 h-7 w-12 rounded-sm' />
          ) : (
            <p className='text-2xl font-bold leading-snug tracking-tight text-white'>
              {value}
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
