'use client'

import { PeerStats } from '@/schema/peers.ts'
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ClockFadingIcon,
  GaugeIcon,
  WifiHighIcon,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'

interface StatsCardProps {
  isLoading: boolean
  stats: PeerStats | undefined
}

function remainingDays(expireTime: string | null | undefined): number {
  if (!expireTime) return 0
  const expireDate = new Date(expireTime)
  const now = new Date()
  const diffTime = expireDate.getTime() - now.getTime()
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
}

export default function PeerStatsCard({ isLoading, stats }: StatsCardProps) {
  // calculate remaining days

  return (
    <Card className='gap-3'>
      <CardHeader>
        <CardTitle>Statistics</CardTitle>
      </CardHeader>

      <CardContent className='flex-1'>
        {isLoading ? (
          <div className='space-y-3'>
            <Skeleton className='h-4 w-1/2' />
            <Skeleton className='h-4 w-1/3' />
            <Skeleton className='h-4 w-2/3' />
            <Skeleton className='h-4 w-1/4' />
            <Skeleton className='h-3 w-full rounded-full' />
          </div>
        ) : (
          <div className='space-y-4 text-sm'>
            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <WifiHighIcon className='h-5 w-5' />
                Traffic Limit
              </span>
              <span>
                {stats?.traffic_limit
                  ? `${stats.traffic_limit} GB`
                  : 'Unlimited'}
              </span>
            </div>

            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <ClockFadingIcon className='h-4 w-4' />
                Expire Time
              </span>
              <span>
                {stats?.expire_time
                  ? `${stats?.expire_time} (${remainingDays(stats?.expire_time)} Days)`
                  : 'Never'}
              </span>
            </div>

            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <ArrowDownIcon className='h-4 w-4' />
                Download
              </span>
              <span>{stats?.download_usage ?? 0} GB</span>
            </div>

            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <ArrowUpIcon className='h-4 w-4' />
                Upload
              </span>
              <span>{stats?.upload_usage ?? 0} GB</span>
            </div>

            <div className='text-foreground flex items-center justify-between font-semibold'>
              <span className='flex items-center gap-2'>
                <GaugeIcon className='h-4 w-4' />
                Total Used
              </span>
              <span>
                {stats?.total_usage ?? 0} GB{' '}
                {stats?.traffic_limit ? `(${stats.usage_percent}%)` : ''}
              </span>
            </div>

            {stats?.traffic_limit && (
              <div>
                <Progress value={Number(stats.usage_percent)} className='h-3' />
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
