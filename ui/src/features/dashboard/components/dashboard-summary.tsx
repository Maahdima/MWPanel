import {
  IconLayersSubtract,
  IconServer,
  IconUserCheck,
  IconUsers,
} from '@tabler/icons-react'
import { DeviceData } from '@/schema/dashboard.ts'
import { HighlightStatsCard } from '@/features/dashboard/components/highlight-stats-card.tsx'
import { StatsCard } from '@/features/dashboard/components/stats-card.tsx'
import { TrafficStatsCard } from '@/features/dashboard/components/traffic-stats-card.tsx'

const countStats = [
  {
    title: 'Total Servers',
    icon: <IconServer />,
    valueKey: 'total_servers',
    source: 'ServerInfo' as const,
  },
  {
    title: 'Total Interfaces',
    icon: <IconLayersSubtract />,
    valueKey: 'total_interfaces',
    source: 'InterfaceInfo' as const,
  },
  {
    title: 'Total Users',
    icon: <IconUsers />,
    valueKey: 'total_peers',
    source: 'PeerInfo' as const,
  },
]

type Props = {
  deviceData: DeviceData | undefined
  isLoading: boolean
}

export function DashboardSummary({ deviceData, isLoading }: Props) {
  return (
    <div className='space-y-4'>
      <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
        {countStats.map((card) => {
          const section = deviceData?.[card.source]
          const value =
            section?.[card.valueKey as keyof typeof section] ?? undefined

          return (
            <StatsCard
              key={card.title}
              title={card.title}
              icon={card.icon}
              value={value}
              isLoading={isLoading}
            />
          )
        })}
      </div>

      <div className='grid gap-4 sm:grid-cols-2'>
        <HighlightStatsCard
          title='Online Users'
          icon={<IconUserCheck />}
          value={deviceData?.PeerInfo?.online_peers}
          isLoading={isLoading}
        />
        <TrafficStatsCard
          value={deviceData?.TrafficInfo?.total_usage}
          isLoading={isLoading}
        />
      </div>
    </div>
  )
}
