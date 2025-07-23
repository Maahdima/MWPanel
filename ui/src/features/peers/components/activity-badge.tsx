import { getAvatarInitials } from '@/utils/helper.ts'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'

interface Props {
  peerName: string
}

export function OnlineBadge({ peerName }: Props) {
  return (
    <div className='relative w-fit'>
      <Avatar className='size-8'>
        <AvatarFallback>{getAvatarInitials(peerName)}</AvatarFallback>
      </Avatar>
      <span className='border-background absolute -end-0.5 -bottom-0.5 size-3 rounded-full border-2 bg-green-600 dark:bg-green-400'>
        <span className='sr-only'>Online</span>
      </span>
    </div>
  )
}

export function OfflineBadge({ peerName }: Props) {
  return (
    <div className='relative w-fit'>
      <Avatar className='size-8'>
        <AvatarFallback>{getAvatarInitials(peerName)}</AvatarFallback>
      </Avatar>
      <span className='border-background absolute -end-0.5 -bottom-0.5 size-3 rounded-full border-2 bg-red-600 dark:bg-red-500'>
        <span className='sr-only'>Offline</span>
      </span>
    </div>
  )
}
