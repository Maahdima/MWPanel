import { JSX } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

type StatsCardProps = {
  title: string
  icon: JSX.Element
  value: number | string | undefined
  isLoading: boolean
}

export function StatsCard({ title, icon, value, isLoading }: StatsCardProps) {
  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
        <CardTitle className='text-lg font-medium'>{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className='mt-1 h-6 w-6 rounded-sm' />
        ) : (
          <div className='text-2xl font-bold'>{value}</div>
        )}
      </CardContent>
    </Card>
  )
}
