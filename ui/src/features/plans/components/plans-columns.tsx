import { ColumnDef } from '@tanstack/react-table'
import { VpnPlan } from '@/schema/vpn-plan.ts'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'
import { PlanRowActions } from '@/features/plans/components/plan-row-actions.tsx'

export const plansColumns: ColumnDef<VpnPlan>[] = [
  {
    id: 'name',
    accessorFn: (plan) => `${plan.name} ${plan.fa_name}`,
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => <div className='text-nowrap'>{row.original.name}</div>,
    meta: { className: cn('pl-3') },
  },
  {
    accessorKey: 'fa_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Bot name' />
    ),
    cell: ({ row }) => (
      <div className='font-iranyekan text-nowrap' dir='auto'>
        {row.original.fa_name}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'traffic_gb',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Traffic' />
    ),
    cell: ({ row }) => <div>{row.original.traffic_gb} GB</div>,
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'duration_days',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Duration' />
    ),
    cell: ({ row }) => <div>{row.original.duration_days} days</div>,
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'price_tomans',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Price' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>
        {row.original.price_tomans.toLocaleString('en-US')} toman
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    id: 'status',
    accessorFn: (plan) => (plan.enabled ? 'enabled' : 'disabled'),
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => (
      <Badge variant={row.original.enabled ? 'default' : 'outline'}>
        {row.original.enabled ? 'enabled' : 'disabled'}
      </Badge>
    ),
    filterFn: (row, columnId, filterValue: string[]) =>
      filterValue.includes(row.getValue(columnId) as string),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'sort_order',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Order' />
    ),
    cell: ({ row }) => <div>{row.original.sort_order}</div>,
    meta: { className: cn('border-l border-r') },
  },
  {
    id: 'actions',
    cell: ({ row }) => <PlanRowActions plan={row.original} />,
    enableSorting: false,
  },
]
