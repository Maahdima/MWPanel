'use client'

import { Skeleton } from '@/components/ui/skeleton'

export function PeerFormSkeleton() {
  return (
    <div className='space-y-4'>
      <Skeleton className='h-6 w-1/3' /> {/* Comment */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Name */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Interface */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Endpoint */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Allowed Address */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Keepalive */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Traffic */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Expire Date */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Bandwidth Download */}
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-6 w-1/3' /> {/* Bandwidth Upload */}
      <Skeleton className='h-10 w-full' />
    </div>
  )
}
