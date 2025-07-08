import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export default function DeviceStatsSkeleton({
  type,
}: {
  type?: 'resource' | 'base'
}) {
  return (
    <Card
      className={
        type === 'base'
          ? 'col-span-1 lg:col-span-3'
          : 'col-span-1 lg:col-span-2'
      }
    >
      <CardContent>
        <h2 className='mb-4 text-lg font-semibold'>
          {type === 'base' ? 'Mikrotik Statistics' : 'Hardware Statistics'}
        </h2>
        <div className='space-y-3'>
          {Array.from({ length: 5 }).map((_, idx) => (
            <div
              key={idx}
              className='flex items-start justify-between border-b border-white/10 py-3'
            >
              <span className='text-sm font-medium text-gray-300'>
                <Skeleton className='h-4 w-24' />
              </span>
              <div className='flex flex-col items-end text-sm text-white'>
                <Skeleton className='h-4 w-32' />
                <Skeleton className='mt-1 h-4 w-24 rounded-md' />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
