import {
  IconLayersSubtract,
  IconServer,
  IconUserCheck,
  IconUsers,
} from '@tabler/icons-react'
import { useDeviceDataQuery } from '@/hooks/dashboard/useDeviceDataQuery.ts'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import DeviceInfo from '@/features/dashboard/components/device-info.tsx'
import DeviceResource from '@/features/dashboard/components/device-resource.tsx'
import RecentlyOnlineUsers from '@/features/dashboard/components/online-users.tsx'
import { PeersChart } from '@/features/dashboard/components/peers-chart.tsx'
import OnlineUsersSkeleton from '@/features/dashboard/components/skeletons/online-users.skeleton.tsx'
import DeviceStatsSkeleton from '@/features/dashboard/components/skeletons/statistics.skeleton.tsx'
import { StatsCard } from '@/features/dashboard/components/stats-card.tsx'
import { TrafficChart } from '@/features/dashboard/components/traffic-chart.tsx'

const statsCards = [
  {
    title: 'Total Servers',
    icon: <IconServer />,
    valueKey: 'total_servers',
    source: 'ServerInfo',
  },
  {
    title: 'Total Interfaces',
    icon: <IconLayersSubtract />,
    valueKey: 'total_interfaces',
    source: 'InterfaceInfo',
  },
  {
    title: 'Total Users',
    icon: <IconUsers />,
    valueKey: 'total_peers',
    source: 'PeerInfo',
  },
  {
    title: 'Online Users',
    icon: <IconUserCheck />,
    valueKey: 'online_peers',
    source: 'PeerInfo',
  },
]

export default function Dashboard() {
  const { data: deviceData, isLoading: isDeviceDataLoading } =
    useDeviceDataQuery()

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <h1 className='text-2xl font-bold tracking-tight'>Dashboard</h1>
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Main ===== */}
      <Main>
        <div className='mb-2 flex items-center justify-between space-y-2 pb-2'>
          <h1 className='text-2xl font-bold tracking-tight'>
            Servers/Clients Summary
          </h1>
        </div>
        <Tabs
          orientation='vertical'
          defaultValue='overview'
          className='space-y-4'
        >
          <TabsContent value='overview' className='space-y-4'>
            <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              {statsCards.map((card) => {
                const section =
                  deviceData?.[card.source as keyof typeof deviceData]
                const value =
                  section?.[card.valueKey as keyof typeof section] ?? '-'
                return (
                  <StatsCard
                    key={card.title}
                    title={card.title}
                    icon={card.icon}
                    value={value}
                    isLoading={isDeviceDataLoading}
                  />
                )
              })}
            </div>

            <div className='grid grid-cols-1 gap-4 lg:grid-cols-7'>
              {isDeviceDataLoading ? (
                <DeviceStatsSkeleton type='base' />
              ) : (
                <DeviceInfo stats={deviceData} />
              )}

              {isDeviceDataLoading ? (
                <DeviceStatsSkeleton type='resource' />
              ) : (
                <DeviceResource stats={deviceData?.DeviceInfo} />
              )}

              <PeersChart
                isLoading={isDeviceDataLoading}
                stats={deviceData?.PeerInfo}
              />

              {isDeviceDataLoading ? (
                <OnlineUsersSkeleton />
              ) : (
                <RecentlyOnlineUsers
                  peers={deviceData?.PeerInfo?.recent_online_peers ?? []}
                />
              )}

              <TrafficChart />
            </div>
          </TabsContent>
        </Tabs>
      </Main>
    </>
  )
}
