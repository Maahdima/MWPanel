import { useState } from 'react'
import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import { downloadPrivate } from '@/api/files'
import {
  useBillingBook,
  useResellerPeersQuery,
} from '@/hooks/resellers/useResellerQueries'
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
import { PeersSessionsDialog } from '@/features/peers/components/dialogs/peers-sessions-dialog'
import { CreatePeerButton } from './create-peer-button'
import { MaintenanceButton } from './maintenance-button'
import { PageFrame } from './page-frame'
import { PurchaseButton } from './purchase-button'

export function MyPeers() {
  const [search, setSearch] = useState('')
  const [selectedPeer, setSelectedPeer] = useState<Peer | null>(null)
  const { data: book } = useBillingBook()
  const readOnly =
    book?.state.billing_suspended || book?.state.administrative_suspended
  const peers = useResellerPeersQuery()
  const download = (path: string, name: string) =>
    downloadPrivate(path, name).catch(() => toast.error('Download failed'))
  return (
    <PageFrame
      title='Configurations'
      description='Your configurations and usage'
      actions={
        <div className='flex flex-wrap gap-2'>
          {!readOnly && <CreatePeerButton />}
          <Button
            variant='outline'
            onClick={() =>
              download('/reseller/me/peers/export', 'traffic-report.xlsx')
            }
          >
            Export usage
          </Button>
        </div>
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
                        {p.migration_pending
                          ? 'Migrating'
                          : p.disabled
                            ? 'Disabled'
                            : 'Enabled'}
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
                            variant='ghost'
                            size='sm'
                            onClick={() => setSelectedPeer(p)}
                          >
                            Sessions
                          </Button>
                          {p.is_shared && (
                            <Button variant='ghost' size='sm' asChild>
                              <a
                                href={`${window.location.origin}/share?shareId=${encodeURIComponent(p.uuid)}`}
                                target='_blank'
                                rel='noopener noreferrer'
                              >
                                Public page
                              </a>
                            </Button>
                          )}
                          {!readOnly && (
                            <>
                              <PurchaseButton peer={p} />
                              <MaintenanceButton peer={p} />
                            </>
                          )}
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
                                `${p.name}.jpeg`
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
      {selectedPeer && (
        <PeersSessionsDialog
          open={true}
          onOpenChange={(open) => {
            if (!open) setSelectedPeer(null)
          }}
          currentRow={selectedPeer}
          scope='reseller'
        />
      )}
    </PageFrame>
  )
}
