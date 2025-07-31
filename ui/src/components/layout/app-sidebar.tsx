import { IconHeartFilled, IconRoute } from '@tabler/icons-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'
import { NavGroup } from '@/components/layout/nav-group'
import { sidebarData } from './data/sidebar-data'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { open } = useSidebar()

  return (
    <Sidebar collapsible='icon' variant='floating' {...props}>
      {open && (
        <SidebarHeader className='flex-row items-center justify-center pt-5'>
          <IconRoute />
          <h1 className='text-2xl font-bold tracking-tight'>MWPanel</h1>
        </SidebarHeader>
      )}

      <SidebarContent>
        {sidebarData.navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>

      {open && (
        <SidebarFooter className='flex items-center justify-center p-4 text-white/80 shadow-inner'>
          <SidebarMenu>
            <SidebarMenuItem>
              <div className='flex items-center justify-center gap-1 text-sm text-xs font-normal'>
                <span>Made with</span>
                <IconHeartFilled size='13' />
                <span>by</span>
                <a
                  href='https://github.com/Maahdima'
                  className='text-xs'
                  target='_blank'
                  rel='noopener noreferrer'
                >
                  Maahdima
                </a>
              </div>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      )}

      <SidebarRail />
    </Sidebar>
  )
}
