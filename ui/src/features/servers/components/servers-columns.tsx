import { ColumnDef } from '@tanstack/react-table'
import { Server } from '@/schema/servers.ts'
import { CheckCircle2Icon, LoaderIcon } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useUpdateServerStatusMutation } from '@/hooks/servers/useUpdateServerStatusMutation.ts'
import { Badge } from '@/components/ui/badge.tsx'
import { Switch } from '@/components/ui/switch.tsx'
import LongText from '@/components/long-text'
import { ServersTableRowActions } from '@/features/servers/components/servers-table-row-actions.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

export const serversColumns: ColumnDef<Server>[] = [
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
    cell: ({ row }) => (
      <Badge
        variant='outline'
        className='text-muted-foreground flex gap-1 px-1.5 [&_svg]:size-3'
      >
        {row.original.status === 'available' ? (
          <CheckCircle2Icon className='text-green-500 dark:text-green-400' />
        ) : (
          <LoaderIcon />
        )}
        Available
      </Badge>
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
    cell: ServersTableRowActions,
  },
]
