'use client'

import { AxiosError } from 'axios'
import { useParams } from '@tanstack/react-router'
import { IconRoute } from '@tabler/icons-react'
import { usePeerConfigQuery } from '@/hooks/peers/usePeerConfigQuery.ts'
import { usePeerQRCodeQuery } from '@/hooks/peers/usePeerQRCodeQuery.ts'
import { usePeerStatsQuery } from '@/hooks/peers/usePeerStatsQuery.ts'
import NotFoundError from '@/features/errors/not-found-error.tsx'
import PeerConfigCard from '@/features/share/components/peer-config-card.tsx'
import PeerQRCodeCard from '@/features/share/components/peer-qrcode-card.tsx'
import PeerStatsCard from '@/features/share/components/peer-stats-card.tsx'

export default function PeerShare() {
  const { uuid } = useParams({ strict: false }) as { uuid: string }

  const {
    data: stats,
    error: statsError,
    isLoading: statsLoading,
  } = usePeerStatsQuery(uuid)

  const { data: configBlob, isLoading: configLoading } =
    usePeerConfigQuery(uuid)

  const { data: qrCode, isLoading: qrCodeLoading } = usePeerQRCodeQuery(uuid)

  if (statsError && (statsError as AxiosError)?.response?.status === 404) {
    return <NotFoundError />
  }

  return (
    <div className='max-w-8xl mx-auto space-y-4 p-6'>
      <div className='space-y-3 text-center'>
        <div className='text-primary flex items-center justify-center gap-2'>
          <IconRoute className='h-6 w-6' />
          <h1 className='text-3xl font-bold'>MWPanel</h1>
          <IconRoute className='h-6 w-6' />
        </div>
        <h2 className='text-foreground text-xl font-semibold'>
          Welcome to your panel{stats?.name ? `: ${stats.name}` : ''}
        </h2>
        <p className='text-muted-foreground text-sm'>
          Scan the QR Code with the WireGuard App to add this peer or download
          the config and import it manually.
        </p>
      </div>

      <div className='grid gap-6 lg:grid-cols-2'>
        <PeerQRCodeCard isLoading={qrCodeLoading} qrCode={qrCode} />

        <div className='space-y-6'>
          <PeerConfigCard
            isLoading={configLoading}
            blob={
              configBlob
                ? new Blob([configBlob], { type: 'text/plain' })
                : undefined
            }
            peerName={stats?.name}
          />
          <PeerStatsCard isLoading={statsLoading} stats={stats} />
        </div>
      </div>
    </div>
  )
}
