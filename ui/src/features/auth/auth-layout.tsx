import { IconRoute } from '@tabler/icons-react'

interface Props {
  children: React.ReactNode
}

export default function AuthLayout({ children }: Props) {
  return (
    <div className='bg-primary-foreground flex min-h-svh w-full items-center justify-center px-4 py-8'>
      <div className='flex w-full min-w-0 flex-col justify-center space-y-2 sm:w-[32rem] lg:w-[36rem]'>
        <div className='mb-4 flex items-center justify-center gap-3'>
          <IconRoute className='size-6' />
          <h1 className='text-xl font-medium sm:text-2xl'>MWPanel</h1>
        </div>
        {children}
      </div>
    </div>
  )
}
