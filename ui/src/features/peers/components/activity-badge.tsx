import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

export function OnlineBadge() {
  return (
    <div className='relative w-fit'>
      <Avatar className='size-8'>
        <AvatarImage
          src='https://cdn.shadcnstudio.com/ss-assets/avatar/avatar-6.png'
          alt='Hallie Richards'
        />
        <AvatarFallback>HR</AvatarFallback>
      </Avatar>
      <span className='border-background absolute -end-0.5 -bottom-0.5 size-3 rounded-full border-2 bg-green-600 dark:bg-green-400'>
        <span className='sr-only'>Online</span>
      </span>
    </div>
  )
}

export function OfflineBadge() {
  return (
    <div className='relative w-fit'>
      <Avatar className='size-8'>
        <AvatarImage
          src='https://cdn.shadcnstudio.com/ss-assets/avatar/avatar-6.png'
          alt='Hallie Richards'
        />
        <AvatarFallback>HR</AvatarFallback>
      </Avatar>
      <span className='border-background absolute -end-0.5 -bottom-0.5 size-3 rounded-full border-2 bg-red-600 dark:bg-red-400'>
        <span className='sr-only'>Offline</span>
      </span>
    </div>
  )
}
