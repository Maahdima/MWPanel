import { ReactNode } from 'react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'

export function PageFrame({
  title,
  description,
  actions,
  children,
}: {
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <>
      <Header fixed>
        <h1 className='truncate text-lg font-semibold'>{title}</h1>
        <div className='ml-auto flex shrink-0 items-center gap-2 sm:gap-4'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <div className='mb-2 flex flex-wrap items-center justify-between gap-2 space-y-2'>
          <div className='min-w-0'>
            <h2 className='text-xl font-semibold tracking-tight'>{title}</h2>
            {description && (
              <p className='text-muted-foreground'>{description}</p>
            )}
          </div>
          {actions}
        </div>
        {children}
      </Main>
    </>
  )
}
