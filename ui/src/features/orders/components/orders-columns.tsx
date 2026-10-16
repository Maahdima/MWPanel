import { ColumnDef } from '@tanstack/react-table'
import { VpnOrder } from '@/schema/vpn-order.ts'
import { formatPanelDateTime } from '@/lib/format-date.ts'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge.tsx'
import { OrderAcceptButton } from '@/features/orders/components/order-accept-button.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

function formatWhen(unix?: number | null) {
  return formatPanelDateTime(unix)
}

function formatCard(card: string) {
  const digits = card.replace(/\D/g, '')
  if (digits.length < 4) return card || '—'
  return digits.replace(/(.{4})/g, '$1-').replace(/-$/, '')
}

function formatTomans(amount: number) {
  return `${amount.toLocaleString('en-US')} toman`
}

const statusClass: Record<string, string> = {
  paid: 'border-transparent bg-emerald-600 text-white',
  pending: 'border-transparent bg-amber-500 text-white',
  fulfilling: 'border-transparent bg-sky-600 text-white',
  failed: 'border-transparent bg-destructive text-white',
  cancelled: 'border-transparent bg-muted text-muted-foreground',
  expired: 'border-transparent bg-muted text-muted-foreground',
}

export const ordersColumns: ColumnDef<VpnOrder>[] = [
  {
    id: 'name',
    accessorFn: (order) =>
      [
        order.tracking_code,
        order.config_name,
        order.username,
        order.payer_card,
        order.plan_name,
        order.plan_title,
        order.chat_id,
        order.peer_name,
      ]
        .filter(Boolean)
        .join(' '),
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Config' />
    ),
    cell: ({ row }) => (
      <div className='w-fit text-nowrap'>
        {row.original.peer_name || row.original.config_name || 'N/A'}
      </div>
    ),
    meta: { className: cn('pl-3') },
  },
  {
    accessorKey: 'tracking_code',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tracking' />
    ),
    cell: ({ row }) => (
      <div className='font-mono text-nowrap'>
        {row.original.tracking_code || 'N/A'}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'kind',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Kind' />
    ),
    cell: ({ row }) => <div className='capitalize'>{row.original.kind}</div>,
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'plan_title',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Plan' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>
        {row.original.plan_name || row.original.plan_title || '—'}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'amount_tomans',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Amount' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>
        {formatTomans(row.original.amount_tomans)}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'payer_card',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Payer card' />
    ),
    cell: ({ row }) => (
      <div className='font-mono text-nowrap'>
        {formatCard(row.original.payer_card)}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) => (
      <div className='flex items-center gap-1'>
        <Badge
          className={statusClass[row.original.status] ?? ''}
          variant='outline'
        >
          {row.original.status}
        </Badge>
        {row.original.manual ? <Badge variant='outline'>Manual</Badge> : null}
      </div>
    ),
    filterFn: (row, columnId, filterValue: string[]) =>
      filterValue.includes(row.getValue(columnId) as string),
    meta: { className: cn('border-l border-r') },
    enableHiding: false,
  },
  {
    accessorKey: 'username',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Telegram' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>
        {row.original.username
          ? `@${row.original.username}`
          : row.original.chat_id}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'created_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Created' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>{formatWhen(row.original.created_at)}</div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'paid_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Paid' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>{formatWhen(row.original.paid_at)}</div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'expires_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Pay until' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>{formatWhen(row.original.expires_at)}</div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'peer_expire',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Config expiry' />
    ),
    cell: ({ row }) => (
      <div className='text-nowrap'>{row.original.peer_expire || 'N/A'}</div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    accessorKey: 'failure_reason',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Failure' />
    ),
    cell: ({ row }) => (
      <div
        className='max-w-56 truncate'
        title={row.original.failure_reason ?? ''}
      >
        {row.original.failure_reason || 'N/A'}
      </div>
    ),
    meta: { className: cn('border-l border-r') },
  },
  {
    id: 'review',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Review' />
    ),
    cell: ({ row }) => <OrderAcceptButton order={row.original} />,
    enableSorting: false,
    meta: { className: cn('border-l border-r') },
  },
]
