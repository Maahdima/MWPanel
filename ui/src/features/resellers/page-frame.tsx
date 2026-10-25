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
        <div className='ml-auto flex items-center gap-3'>
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>
      <Main>
        <div className='mb-5 flex flex-wrap items-start justify-between gap-3'>
          <div>
            <h2 className='text-xl font-semibold tracking-tight'>{title}</h2>
            {description && (
              <p className='text-muted-foreground mt-1'>{description}</p>
            )}
          </div>
          {actions}
        </div>
        {children}
      </Main>
    </>
  )
}
