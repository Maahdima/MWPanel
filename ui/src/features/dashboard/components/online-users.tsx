import { IconUserOff } from '@tabler/icons-react'
import { DeviceData } from '@/schema/dashboard.ts'
import { getAvatarInitials } from '@/utils/helper.ts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

type Props = {
  peers: DeviceData['PeerInfo']['recent_online_peers']
}

export default function RecentlyOnlineUsers({ peers }: Props) {
  const hasPeers = peers && peers.length > 0

  return (
    <Card className='col-span-1 flex h-full flex-col lg:col-span-2'>
      <CardHeader>
        <CardTitle className='mb-4 text-lg font-semibold'>
          Recently Online Users
        </CardTitle>
      </CardHeader>
      <CardContent className='flex flex-grow flex-col p-0'>
        {!hasPeers ? (
          <div className='flex flex-grow flex-col items-center justify-center gap-4 text-slate-500'>
            <IconUserOff size={64} stroke={1.5} />
            <p className='text-center text-base'>
              No users have been online recently.
            </p>
          </div>
        ) : (
          <div className='-mt-4'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='pl-6'>User</TableHead>
                  <TableHead className='pr-6 text-right'>Last Seen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {peers.map((peer, idx) => (
                  <TableRow key={idx}>
                    <TableCell className='py-4 pl-6 font-medium'>
                      <div className='flex items-center gap-3'>
                        <div className='flex h-9 w-9 items-center justify-center rounded-full bg-teal-800 text-sm font-bold text-white'>
                          {getAvatarInitials(peer.name)}
                        </div>
                        <span>{peer.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className='pr-6 text-right text-slate-400'>
                      {peer.last_seen} ago
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
