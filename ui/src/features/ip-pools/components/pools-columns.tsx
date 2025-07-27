import { ColumnDef } from '@tanstack/react-table'
import { IPPool } from '@/schema/ip-pool.ts'
import { cn } from '@/lib/utils'
import { PoolsTableRowActions } from '@/features/ip-pools/components/pools-table-row-actions.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

export const poolsColumns: ColumnDef<IPPool>[] = [
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
      className: cn('pl-3'),
    },
  },
  {
    accessorKey: 'start_ip',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Start IP' />
    ),
    cell: ({ row }) => {
      const { start_ip } = row.original
      return <div className='w-fit text-nowrap'>{start_ip}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
  },
  {
    accessorKey: 'end_ip',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='End IP' />
    ),
    cell: ({ row }) => {
      const { end_ip } = row.original
      return <div className='w-fit text-nowrap'>{end_ip}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
  },
  {
    accessorKey: 'total_ip',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Total IPs' />
    ),
    cell: ({ row }) => {
      const { total_ip } = row.original
      return <div className='w-fit text-nowrap'>{total_ip}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'used_ip',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Used IPs' />
    ),
    cell: ({ row }) => {
      const { used_ip } = row.original
      return <div className='w-fit text-nowrap'>{used_ip}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'remaining_ip',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Remaining IPs' />
    ),
    cell: ({ row }) => {
      const { remaining_ip } = row.original
      return <div className='w-fit text-nowrap'>{remaining_ip}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    id: 'actions',
    cell: PoolsTableRowActions,
  },
]
