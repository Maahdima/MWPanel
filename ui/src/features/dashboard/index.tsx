import { useDeviceDataQuery } from '@/hooks/dashboard/useDeviceDataQuery.ts'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { DashboardSummary } from '@/features/dashboard/components/dashboard-summary.tsx'
import DeviceInfo from '@/features/dashboard/components/device-info.tsx'
import DeviceResource from '@/features/dashboard/components/device-resource.tsx'
import RecentlyOnlineUsers from '@/features/dashboard/components/online-users.tsx'
import { PeersChart } from '@/features/dashboard/components/peers-chart.tsx'
import OnlineUsersSkeleton from '@/features/dashboard/components/skeletons/online-users.skeleton.tsx'
import DeviceStatsSkeleton from '@/features/dashboard/components/skeletons/statistics.skeleton.tsx'
import { TrafficChart } from '@/features/dashboard/components/traffic-chart.tsx'

export default function Dashboard() {
  const { data: deviceData, isLoading: isDeviceDataLoading } =
    useDeviceDataQuery()

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <h1 className='hidden truncate text-lg font-semibold tracking-tight sm:block'>
          Dashboard
        </h1>
        <div className='ml-auto flex min-w-0 items-center gap-2 sm:gap-4'>
          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Main ===== */}
      <Main>
        <div className='mb-2 flex flex-wrap items-center justify-between gap-2 pb-2'>
          <h1 className='text-xl font-semibold tracking-tight'>
            Servers/Clients Summary
          </h1>
        </div>
        <Tabs
          orientation='vertical'
          defaultValue='overview'
          className='space-y-4'
        >
          <TabsContent value='overview' className='space-y-4'>
            <DashboardSummary
              deviceData={deviceData}
              isLoading={isDeviceDataLoading}
            />

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
