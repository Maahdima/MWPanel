'use client'

import { IconBrandTelegram } from '@tabler/icons-react'
import { BellIcon, UserIcon } from 'lucide-react'
import { ColoredBadge } from '@/features/shared-components/status-badge.tsx'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import type { PeerTelegramStatus, TelegramStatus } from '@/schema/telegram.ts'

interface PeerTelegramCardProps {
  isLoading: boolean
  shareId: string
  botStatus?: TelegramStatus
  linkStatus?: PeerTelegramStatus
}

export default function PeerTelegramCard({
  isLoading,
  shareId,
  botStatus,
  linkStatus,
}: PeerTelegramCardProps) {
  const startURL = botStatus?.bot_url
    ? `${botStatus.bot_url}?start=${shareId}`
    : undefined
  const linked = Boolean(linkStatus?.linked)
  const alertsOn = Boolean(linkStatus?.notify_enabled)
  let alertsLabel = '—'
  if (linked) {
    alertsLabel = alertsOn ? 'On' : 'Off'
  }
  const botName = botStatus?.bot_username
    ? `@${botStatus.bot_username.replace(/^@/, '')}`
    : '—'
  const account = linkStatus?.username || (linked ? 'Chat linked' : '—')

  return (
    <Card className='flex h-full flex-col gap-3'>
      <CardHeader className='flex flex-row items-center justify-between space-y-0'>
        <CardTitle>Telegram</CardTitle>
        {!isLoading && (
          <ColoredBadge
            color={linked ? 'green' : 'yellow'}
            text={linked ? 'linked' : 'not linked'}
          />
        )}
      </CardHeader>

      <CardContent className='flex flex-1 flex-col justify-center'>
        {isLoading ? (
          <div className='space-y-3'>
            <Skeleton className='h-4 w-1/2' />
            <Skeleton className='h-4 w-2/3' />
            <Skeleton className='h-4 w-1/3' />
          </div>
        ) : (
          <div className='space-y-3 text-sm'>
            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <IconBrandTelegram className='h-4 w-4' />
                Bot
              </span>
              <span>{botName}</span>
            </div>

            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <UserIcon className='h-4 w-4' />
                Account
              </span>
              <span>{account}</span>
            </div>

            <div className='flex items-center justify-between'>
              <span className='flex items-center gap-2'>
                <BellIcon className='h-4 w-4' />
                Alerts
              </span>
              <span>{alertsLabel}</span>
            </div>
          </div>
        )}
      </CardContent>

      {startURL && (
        <CardFooter className='mt-auto'>
          <a
            href={startURL}
            target='_blank'
            rel='noreferrer'
            className='bg-[#2AABEE] hover:bg-[#229ED9] inline-flex w-full items-center justify-center gap-2 rounded-md px-5 py-2.5 text-sm font-semibold text-white shadow-md transition-colors'
          >
            <IconBrandTelegram className='h-5 w-5' />
            {linked ? 'Open Telegram bot' : 'Connect to Telegram bot'}
          </a>
        </CardFooter>
      )}
    </Card>
  )
}
