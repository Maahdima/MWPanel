'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface QRCodeCardProps {
  isLoading: boolean
  qrCode?: string
}

export default function PeerQRCodeCard({ isLoading, qrCode }: QRCodeCardProps) {
  const [isBlurred, setIsBlurred] = useState(true)

  const handleToggleBlur = () => {
    setIsBlurred((prev) => !prev)
  }

  return (
    <Card className='flex h-full flex-col'>
      <CardHeader>
        <CardTitle>QR Code</CardTitle>
      </CardHeader>

      <CardContent className='flex flex-1 items-center justify-center'>
        {isLoading ? (
          <Skeleton className='h-[300px] w-[300px] rounded-md' />
        ) : (
          <div
            onClick={handleToggleBlur}
            className='relative cursor-pointer'
            title={isBlurred ? 'Click to reveal' : 'Click to hide'}
          >
            <img
              src={qrCode}
              alt='QR Code'
              width={300}
              height={300}
              className={`rounded-md transition-all duration-300 ${
                isBlurred ? 'blur-md' : 'blur-0'
              }`}
            />
            {isBlurred && (
              <div className='absolute inset-0 flex items-center justify-center rounded-md bg-black/40 font-semibold text-white'>
                Click to reveal
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
