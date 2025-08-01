import { useState } from 'react'
import { ColumnDef } from '@tanstack/react-table'
import { IconRestore } from '@tabler/icons-react'
import { Peer, PeerStatus } from '@/schema/peers.ts'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useResetUsageMutation } from '@/hooks/peers/useResetUsageMutation.ts'
import { useUpdatePeerStatusMutation } from '@/hooks/peers/useUpdatePeerStatusMutation.ts'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button.tsx'
import { Switch } from '@/components/ui/switch.tsx'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import LongText from '@/components/long-text'
import {
  OfflineBadge,
  OnlineBadge,
} from '@/features/peers/components/activity-badge.tsx'
import { PeersTableRowActions } from '@/features/peers/components/peers-table-row-actions.tsx'
import { ColoredBadge } from '@/features/shared-components/status-badge.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'
import { SimpleDialog } from '@/features/shared-components/table/dialogs/simple-dialog.tsx'

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
    id: 'name',
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name, is_online } = row.original
      return (
        <div className='flex w-fit items-center justify-center gap-3 text-nowrap'>
          {is_online ? (
            <OnlineBadge peerName={name} />
          ) : (
            <OfflineBadge peerName={name} />
          )}
          {name}
        </div>
      )
    },
    meta: {
      className: cn('border-l border-r'),
    },
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
      const [dialogOpen, setDialogOpen] = useState(false)

      const resetUsageMutation = useResetUsageMutation()

      const peer = row.original

      const handleResetUsage = async () => {
        resetUsageMutation.mutateAsync(peer.id, {
          onSuccess: () => {
            setDialogOpen(false)
            toast.success('Peer usage reset successfully', {
              duration: 5000,
            })
          },
          onError: () => {
            setDialogOpen(false)
          },
        })
      }

      return (
        <SimpleDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          title='Reset Peer Usage?'
          description='This will reset the peer’s usage statistics. Are you sure?'
          actionText='Confirm Reset'
          mutateAsync={handleResetUsage}
          trigger={
            <div className='min-w-[140px] text-nowrap'>
              <div className='flex items-center gap-3'>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant='ghost'
                      size='icon'
                      className='text-muted-foreground hover:text-foreground h-8 w-8'
                      aria-label='Reset Usage'
                    >
                      <IconRestore className='h-4 w-4' />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side='top' align='center'>
                    <p>Reset Usage</p>
                  </TooltipContent>
                </Tooltip>
                <div className='flex flex-col leading-tight'>
                  <span className='text-foreground text-sm font-medium'>
                    {peer.total_usage} GB
                    {peer.traffic_limit && (
                      <span className='text-muted-foreground text-sm'>
                        {' / '}
                        {peer.traffic_limit} GB
                      </span>
                    )}
                  </span>
                  {!peer.traffic_limit && (
                    <span className='text-muted-foreground text-xs'>
                      Unlimited
                    </span>
                  )}
                </div>
              </div>
            </div>
          }
        />
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
          {expire_time ? expire_time : 'Never'}
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
        <div className='w-fit min-w-[150px] text-nowrap'>
          {row.original ? (
            <div className='text-foreground flex items-center gap-2 text-sm'>
              <div className='flex items-center gap-1'>
                <span className='text-muted-foreground'>↓</span>
                <span>{download_bandwidth || 'Unlimited'}</span>
              </div>
              <span className='text-muted-foreground'>/</span>
              <div className='flex items-center gap-1'>
                <span className='text-muted-foreground'>↑</span>
                <span>{upload_bandwidth || 'Unlimited'}</span>
              </div>
            </div>
          ) : (
            <Badge
              variant='outline'
              className='text-muted-foreground rounded-sm text-xs'
            >
              Unlimited
            </Badge>
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
            ? status.map((status: PeerStatus, idx: number) => (
                <ColoredBadge
                  key={idx}
                  color={
                    status === 'active'
                      ? 'green'
                      : status === 'inactive'
                        ? 'gray'
                        : status === 'expired'
                          ? 'yellow'
                          : 'red'
                  }
                  text={status}
                />
              ))
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
