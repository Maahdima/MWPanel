import {
  IconCloud,
  IconHelp,
  IconLayoutDashboard,
  IconLink,
  IconPalette,
  IconSettings,
  IconTool,
  IconUsers,
  IconWorld,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  // user: {
  //   name: 'maahdima',
  //   email: 'maahdima@gmail.com',
  //   avatar: '/avatars/shadcn.jpg',
  // },
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: IconLayoutDashboard,
        },
        {
          title: 'Servers',
          url: '/servers',
          icon: IconCloud,
        },
        {
          title: 'Interfaces',
          url: '/interfaces',
          icon: IconWorld,
        },
        {
          title: 'Pools',
          url: '/pools',
          icon: IconLink,
        },
        {
          title: 'Peers',
          url: '/peers',
          icon: IconUsers,
        },
      ],
    },
    {
      title: 'Other',
      items: [
        {
          title: 'Settings',
          icon: IconSettings,
          items: [
            {
              title: 'Account',
              url: '/settings/account',
              icon: IconTool,
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
              icon: IconPalette,
            },
          ],
        },
        {
          title: 'Help Center',
          url: '/help-center',
          icon: IconHelp,
        },
      ],
    },
  ],
}
