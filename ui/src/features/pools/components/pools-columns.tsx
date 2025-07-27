import { ColumnDef } from '@tanstack/react-table'
import { Server } from '@/schema/servers.ts'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useUpdateServerStatusMutation } from '@/hooks/servers/useUpdateServerStatusMutation.ts'
import { Switch } from '@/components/ui/switch.tsx'
import LongText from '@/components/long-text'
import { PoolsTableRowActions } from '@/features/pools/components/pools-table-row-actions.tsx'
import { ColoredBadge } from '@/features/shared-components/status-badge.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

export const poolsColumns: ColumnDef<Server>[] = [
  {
    id: 'is_active',
    cell: ({ row }) => {
      const server = row.original
      const updateServerStatusMutation = useUpdateServerStatusMutation()

      const handleToggle = () => {
        updateServerStatusMutation.mutate(server.id, {
          onSuccess: () => {
            toast.success(
              `Server ${server.is_active ? 'disabled' : 'enabled'} successfully`,
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
            id={`status-${server.id}`}
            checked={server.is_active}
            onCheckedChange={handleToggle}
            disabled={updateServerStatusMutation.isPending}
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
      const { name } = row.original
      return <div className='w-fit text-nowrap'>{name}</div>
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
    accessorKey: 'ip_address',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='IP Address' />
    ),
    cell: ({ row }) => <div>{row.getValue('ip_address')}</div>,
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'api_port',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='API Port' />
    ),
    cell: ({ row }) => <div>{row.getValue('api_port')}</div>,
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    id: 'status',
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) =>
      row.original.status === 'available' ? (
        <ColoredBadge color='green' text='Available' />
      ) : (
        <ColoredBadge color='red' text='Unavailable' />
      ),
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
    cell: PoolsTableRowActions,
  },
]
