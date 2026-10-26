import { useMemo, useState } from 'react'
import { ColumnDef } from '@tanstack/react-table'
import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import { downloadPrivate } from '@/api/files'
import {
  useBillingBook,
  useResellerPeersQuery,
} from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import { PeersSessionsDialog } from '@/features/peers/components/dialogs/peers-sessions-dialog'
import { DataTable } from '@/features/shared-components/table/data-table'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton'
import { CreatePeerButton } from './components/create-peer-button'
import { MaintenanceButton } from './components/maintenance-button'
import { PurchaseButton } from './components/purchase-button'
import { PageFrame } from './page-frame'

const download = (path: string, name: string) =>
  downloadPrivate(path, name).catch(() => toast.error('Download failed'))

export function MyPeers() {
  const [selectedPeer, setSelectedPeer] = useState<Peer | null>(null)
  const { data: book } = useBillingBook()
  const readOnly =
    book?.state.billing_suspended || book?.state.administrative_suspended
  const peers = useResellerPeersQuery()
  const columns = useMemo<ColumnDef<Peer>[]>(
    () => [
      {
        id: 'name',
        accessorFn: (peer) => `${peer.name} ${peer.comment ?? ''}`,
        header: 'Name',
        cell: ({ row }) => row.original.name,
      },
      {
        id: 'status',
        accessorKey: 'status',
        filterFn: 'arrIncludesSome',
        enableSorting: false,
        header: 'Status',
        cell: ({ row }) => {
          const p = row.original
          return (
            <>
              {p.migration_pending
                ? 'Migrating'
                : p.disabled
                  ? 'Disabled'
                  : 'Enabled'}
              <div className='text-muted-foreground text-xs'>
                {p.restriction_reasons?.join(', ')}
                {p.sync_status ? ` · ${p.sync_status}` : ''}
              </div>
            </>
          )
        },
      },
      { accessorKey: 'total_usage', header: 'Usage' },
      {
        accessorKey: 'traffic_limit',
        header: 'Quota',
        cell: ({ row }) => row.original.traffic_limit ?? 'Unlimited',
      },
      {
        accessorKey: 'expire_time',
        header: 'Expires',
        cell: ({ row }) => row.original.expire_time ?? 'Unlimited',
      },
      {
        id: 'actions',
        header: 'Actions',
        cell: ({ row }) => {
          const p = row.original
          return (
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
          )
        },
      },
    ],
    [readOnly]
  )
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
      {peers.isPending ? (
        <DataTableSkeleton columns={6} />
      ) : peers.isError ? (
        <p role='alert'>
          Unable to load configurations.{' '}
          <Button onClick={() => peers.refetch()}>Retry</Button>
        </p>
      ) : (
        <DataTable data={peers.data ?? []} columns={columns} type='Peers' />
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
