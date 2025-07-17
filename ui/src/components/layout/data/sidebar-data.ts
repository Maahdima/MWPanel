import {
  IconCloudComputing,
  IconHelp,
  IconLayoutDashboard,
  IconPalette,
  IconSettings,
  IconTool,
  IconTransferVertical,
  IconUsers,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'maahdima',
    email: 'maahdima@gmail.com',
    avatar: '/avatars/shadcn.jpg',
  },
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
          icon: IconCloudComputing,
        },
        {
          title: 'Interfaces',
          url: '/interfaces',
          icon: IconTransferVertical,
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
