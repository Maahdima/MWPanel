import { IconRoute } from '@tabler/icons-react'

interface Props {
  children: React.ReactNode
}

export default function AuthLayout({ children }: Props) {
  return (
    <div className='bg-primary-foreground flex min-h-svh w-full items-center justify-center px-4 py-8'>
      <div className='flex w-full max-w-[26.25rem] flex-col justify-center space-y-3'>
        <div className='mb-1 flex items-center justify-center gap-2.5'>
          <IconRoute className='size-5' />
          <h1 className='text-lg font-medium tracking-tight'>MWPanel</h1>
        </div>
        {children}
      </div>
    </div>
  )
}
