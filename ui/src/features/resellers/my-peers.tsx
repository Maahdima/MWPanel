import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { PeersResponseSchema } from '@/schema/peers'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import { downloadPrivate } from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableHeader,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
} from '@/components/ui/table'
import { PageFrame } from './page-frame'

export function MyPeers() {
  const [search, setSearch] = useState('')
  const peers = useQuery({
    queryKey: ['my_peers'],
    queryFn: async () =>
      PeersResponseSchema.parse(
        (await axiosInstance.get('/reseller/me/peers')).data
      ).data,
    refetchInterval: 15000,
  })
  const download = (path: string, name: string) =>
    downloadPrivate(path, name).catch(() => toast.error('Download failed'))
  return (
    <PageFrame
      title='Configurations'
      description='Your configurations and usage'
      actions={
        <Button
          variant='outline'
          onClick={() =>
            download('/reseller/me/peers/export', 'traffic-report.xlsx')
          }
        >
          Export usage
        </Button>
      }
    >
      <Input
        aria-label='Search configurations'
        placeholder='Search configurations…'
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className='mb-4 max-w-sm'
      />
      {peers.isPending ? (
        <p>Loading configurations…</p>
      ) : peers.isError ? (
        <p>
          Unable to load configurations.{' '}
          <Button onClick={() => peers.refetch()}>Retry</Button>
        </p>
      ) : (
        <Card>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  {[
                    'Name',
                    'Status',
                    'Usage',
                    'Quota',
                    'Expires',
                    'Actions',
                  ].map((x) => (
                    <TableHead key={x}>{x}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {(peers.data ?? [])
                  .filter((p) =>
                    `${p.name} ${p.comment ?? ''}`
                      .toLowerCase()
                      .includes(search.toLowerCase())
                  )
                  .map((p) => (
                    <TableRow key={p.id}>
                      <TableCell>{p.name}</TableCell>
                      <TableCell>
                        {p.disabled ? 'Disabled' : 'Enabled'}
                        <div className='text-muted-foreground text-xs'>
                          {p.restriction_reasons?.join(', ')}
                          {p.sync_status ? ` · ${p.sync_status}` : ''}
                        </div>
                      </TableCell>
                      <TableCell>{p.total_usage}</TableCell>
                      <TableCell>{p.traffic_limit ?? 'Unlimited'}</TableCell>
                      <TableCell>{p.expire_time ?? 'Unlimited'}</TableCell>
                      <TableCell>
                        <div className='flex flex-wrap gap-2'>
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() =>
                              download(
                                `/reseller/me/peers/${p.id}/config`,
                                `${p.name}.conf`
                              )
                            }
                          >
                            Config
                          </Button>
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() =>
                              download(
                                `/reseller/me/peers/${p.id}/qrcode`,
                                `${p.name}.png`
                              )
                            }
                          >
                            QR code
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                {!peers.data?.length && (
                  <TableRow>
                    <TableCell colSpan={6}>
                      No configurations assigned.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </PageFrame>
  )
}
