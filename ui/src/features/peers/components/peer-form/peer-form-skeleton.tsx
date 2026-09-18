'use client'

import { Skeleton } from '@/components/ui/skeleton'

export function PeerFormSkeleton() {
  return (
    <div className='grid grid-cols-1 gap-4 @min-[28rem]:grid-cols-2'>
      <div className='space-y-2 @min-[28rem]:col-span-2'>
        <Skeleton className='h-4 w-16' />
        <Skeleton className='h-9 w-full' />
      </div>
      <div className='space-y-2 @min-[28rem]:col-span-2'>
        <Skeleton className='h-4 w-20' />
        <Skeleton className='h-9 w-full' />
      </div>
      {Array.from({ length: 6 }).map((_, index) => (
        <div key={index} className='space-y-2'>
          <Skeleton className='h-4 w-24' />
          <Skeleton className='h-9 w-full' />
        </div>
      ))}
    </div>
  )
}
