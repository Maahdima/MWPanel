'use client'

import { AxiosError } from 'axios'
import { useSearch } from '@tanstack/react-router'
import { IconRoute } from '@tabler/icons-react'
import { useTelegramStatusQuery } from '@/hooks/telegram/useTelegramStatusQuery'
import { usePeerTelegramStatusQuery } from '@/hooks/telegram/usePeerTelegramStatusQuery.ts'
import { useUserConfigQuery } from '@/hooks/user/useUserConfigQuery.ts'
import { useUserDetailsQuery } from '@/hooks/user/useUserDetailsQuery.ts'
import { useUserQRCodeQuery } from '@/hooks/user/useUserQRCodeQuery.ts'
import NotFoundError from '@/features/errors/not-found-error.tsx'
import PeerConfigCard from '@/features/share/components/peer-config-card.tsx'
import PeerQRCodeCard from '@/features/share/components/peer-qrcode-card.tsx'
import PeerStatsCard from '@/features/share/components/peer-stats-card.tsx'
import PeerTelegramCard from '@/features/share/components/peer-telegram-card.tsx'

export default function PeerShare() {
  const { shareId } = useSearch({ from: '/share' })

  const {
    data: stats,
    error: statsError,
    isLoading: statsLoading,
  } = useUserDetailsQuery(shareId)

  const { data: configBlob, isLoading: configLoading } =
    useUserConfigQuery(shareId)

  const { data: qrCode, isLoading: qrCodeLoading } = useUserQRCodeQuery(shareId)
  const { data: telegramStatus, isLoading: telegramStatusLoading } =
    useTelegramStatusQuery()
  const { data: telegramLink, isLoading: telegramLinkLoading } =
    usePeerTelegramStatusQuery(
      telegramStatus?.enabled ? shareId : undefined
    )

  const configCard = (
    <PeerConfigCard
      isLoading={configLoading}
      blob={
        configBlob
          ? new Blob([configBlob], { type: 'text/plain' })
          : undefined
      }
      peerName={stats?.name}
    />
  )

  const statsCard = (
    <PeerStatsCard isLoading={statsLoading} stats={stats} />
  )

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

      {telegramStatus?.enabled ? (
        <div className='grid gap-6 lg:grid-cols-2'>
          <PeerQRCodeCard isLoading={qrCodeLoading} qrCode={qrCode} />
          {configCard}
          <PeerTelegramCard
            isLoading={telegramStatusLoading || telegramLinkLoading}
            shareId={shareId}
            botStatus={telegramStatus}
            linkStatus={telegramLink}
          />
          {statsCard}
        </div>
      ) : (
        <div className='grid items-start gap-6 lg:grid-cols-2'>
          <PeerQRCodeCard isLoading={qrCodeLoading} qrCode={qrCode} />
          <div className='space-y-6'>
            {configCard}
            {statsCard}
          </div>
        </div>
      )}
    </div>
  )
}
