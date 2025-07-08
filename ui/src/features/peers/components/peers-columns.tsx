import { ColumnDef } from '@tanstack/react-table'
import { Peer, PeerStatus } from '@/schema/peers.ts'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useUpdatePeerStatusMutation } from '@/hooks/peers/useUpdatePeerStatusMutation.ts'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch.tsx'
import LongText from '@/components/long-text'
import { PeersTableRowActions } from '@/features/peers/components/peers-table-row-actions.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

const statusClass = new Map<PeerStatus, string>([
  ['active', 'bg-teal-100/30 text-teal-900 dark:text-teal-200 border-teal-200'],
  ['inactive', 'bg-neutral-300/40 dark:text-neutral-100 border-neutral-300'],
  [
    'expired',
    'bg-yellow-500/40 text-yellow-900 dark:text-yellow-100 border-yellow-300',
  ],
  [
    'suspended',
    'bg-destructive/10 dark:bg-destructive/50 text-destructive dark:text-primary border-destructive/10',
  ],
])

export const peersColumns: ColumnDef<Peer>[] = [
  {
    id: 'is_active',
    cell: ({ row }) => {
      const peer = row.original
      const updatePeerStatusMutation = useUpdatePeerStatusMutation()

      const handleToggle = () => {
        updatePeerStatusMutation.mutate(peer.id, {
          onSuccess: () => {
            toast.success(
              `Peer ${!peer.disabled ? 'disabled' : 'enabled'} successfully`,
              {
                duration: 5000,
              }
            )
          },
        })
      }

      return (
        <div className='flex items-center justify-center'>
          <Switch
            id={`status-${peer.id}`}
            checked={!peer.disabled}
            onCheckedChange={handleToggle}
            disabled={updatePeerStatusMutation.isPending}
          />
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Comment' />
    ),
    cell: ({ row }) => (
      <LongText className='max-w-36'>
        {row.getValue('comment') ?? (
          <span className='text-muted-foreground'>N/A</span>
        )}
      </LongText>
    ),
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
  },
  {
    id: 'name',
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name } = row.original
      return <div className='w-fit text-nowrap'>{name}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'interface',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Interface' />
    ),
    cell: ({ row }) => (
      <div className='w-fit text-nowrap'>{row.getValue('interface')}</div>
    ),
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'allowed_address',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='IP Address' />
    ),
    cell: ({ row }) => <div>{row.getValue('allowed_address')}</div>,
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'traffic_limit',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Traffic' />
    ),
    cell: ({ row }) => {
      const { traffic_limit } = row.original
      return (
        <div className='w-fit text-nowrap'>
          {traffic_limit ? (
            <Badge variant='default' className='bg-blue-400'>
              {traffic_limit} GB
            </Badge>
          ) : (
            <Badge variant='default'>Unlimited</Badge>
          )}
        </div>
      )
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'expire',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Expire' />
    ),
    cell: ({ row }) => {
      const { expire_time } = row.original
      return (
        <div className='w-fit text-nowrap'>
          {expire_time ? (
            <Badge variant='default' className='bg-yellow-400'>
              {expire_time}
            </Badge>
          ) : (
            <Badge variant='default'>Unlimited</Badge>
          )}
        </div>
      )
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'bandwidth',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Bandwidth' />
    ),
    cell: ({ row }) => {
      const { download_bandwidth, upload_bandwidth } = row.original
      return (
        <div className='w-fit text-nowrap'>
          {row.original ? (
            <Badge variant='default' className='bg-purple-500'>
              {download_bandwidth || 'Unlimited'}/
              {upload_bandwidth || 'Unlimited'}
            </Badge>
          ) : (
            <Badge variant='default'>Unlimited</Badge>
          )}
        </div>
      )
    },
    meta: {
      className: cn('border-l border-r'),
    },
    enableSorting: true,
  },
  {
    id: 'status',
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => {
      const { status } = row.original
      return (
        <div className='flex space-x-2'>
          {Array.isArray(status)
            ? status.map((status: PeerStatus, idx: number) => {
                const badgeColor = statusClass.get(status)
                return (
                  <Badge key={idx} className={cn('capitalize', badgeColor)}>
                    {status}
                  </Badge>
                )
              })
            : null}
        </div>
      )
    },
    filterFn: (row, columnId, filterValue: string[]) => {
      const cellValue = row.getValue(columnId) as string[]
      return filterValue.some((val) => cellValue.includes(val))
    },
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: PeersTableRowActions,
  },
]
