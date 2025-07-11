'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface QRCodeCardProps {
  isLoading: boolean
  qrCode?: string
}

export default function PeerQRCodeCard({ isLoading, qrCode }: QRCodeCardProps) {
  return (
    <Card className='flex h-full flex-col'>
      <CardHeader>
        <CardTitle>QR Code</CardTitle>
      </CardHeader>

      <CardContent className='flex flex-1 items-center justify-center'>
        {isLoading ? (
          <Skeleton className='h-[300px] w-[300px] rounded-md' />
        ) : (
          <img
            src={qrCode}
            alt='QR Code'
            width={450}
            height={450}
            className='rounded-md'
          />
        )}
      </CardContent>
    </Card>
  )
}
