import { IconUserOff } from '@tabler/icons-react'
import { DeviceData } from '@/schema/dashboard.ts'
import { Card, CardContent } from '@/components/ui/card'

type RecentlyOnlineUsersProps = {
  peers: DeviceData['PeerInfo']['recent_online_peers']
}
export default function RecentlyOnlineUsers({
  peers,
}: RecentlyOnlineUsersProps) {
  return (
    <Card className='col-span-1 lg:col-span-2'>
      <CardContent className='h-full px-6'>
        <h2 className='mb-4 border-white/10 pb-2 text-lg font-semibold'>
          Recently Online Users
        </h2>
        {!peers || peers.length === 0 ? (
          <div className='flex h-[77%] flex-col items-center justify-center text-gray-400'>
            <IconUserOff className='pb-5' size={80} />
            <p className='text-center text-lg'>
              No users have been online recently.
            </p>
          </div>
        ) : (
          <table className='w-full text-sm'>
            <thead>
              <tr className='text-center text-gray-400'>
                <th className='justify-center py-2'>Name</th>
                <th className='justify-center py-2'>Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {peers?.map((peer, idx) => (
                <tr
                  key={idx}
                  className='justify-center border-t border-white/5 transition hover:bg-white/5'
                >
                  <td className='flex items-center justify-center gap-3 py-4'>
                    <div
                      className={`flex h-8 w-8 items-center justify-center rounded-full bg-yellow-500 text-xs font-bold text-white`}
                    >
                      {peer?.name.split('-')[1]?.slice(0, 2).toUpperCase() ||
                        'U'}
                    </div>
                    <span>{peer?.name}</span>
                  </td>
                  <td className='py-2 text-gray-300'>{peer?.last_seen} ago</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </CardContent>
    </Card>
  )
}
