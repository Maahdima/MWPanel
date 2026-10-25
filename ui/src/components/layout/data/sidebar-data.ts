import {
  IconCloud,
  IconHelp,
  IconLayoutDashboard,
  IconLink,
  IconPackage,
  IconPalette,
  IconReceipt,
  IconSettings,
  IconShieldLock,
  IconTool,
  IconUsers,
  IconWorld,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'

const shopNavUrls = new Set<string>(['/plans', '/orders'])

export function sidebarDataForShop(enabled: boolean): SidebarData {
  if (enabled) {
    return sidebarData
  }

  return {
    navGroups: sidebarData.navGroups.map((group) => ({
      ...group,
      items: group.items.filter(
        (item) => !('url' in item && item.url && shopNavUrls.has(item.url))
      ),
    })),
  }
}

export const sidebarData: SidebarData = {
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
          title: 'Resellers',
          url: '/resellers',
          icon: IconUsers,
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
        {
          title: 'Plans',
          url: '/plans',
          icon: IconPackage,
        },
        {
          title: 'Orders',
          url: '/orders',
          icon: IconReceipt,
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
              title: 'Security',
              url: '/settings/security',
              icon: IconShieldLock,
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
