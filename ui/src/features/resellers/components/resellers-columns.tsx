import { Link } from '@tanstack/react-router'
import { ColumnDef } from '@tanstack/react-table'
import { Reseller } from '@/schema/billing'
import { Badge } from '@/components/ui/badge'
import { dateLabel, formatToman } from '../money'

export const resellersColumns: ColumnDef<Reseller>[] = [
  {
    id: 'name',
    accessorFn: (row) => `${row.name} ${row.username}`,
    header: 'Reseller',
    cell: ({ row }) => (
      <Link
        to='/resellers/$id'
        params={{ id: String(row.original.id) }}
        className='font-medium underline-offset-4 hover:underline'
      >
        {row.original.name}
        <span className='text-muted-foreground block text-xs'>
          {row.original.username}
        </span>
      </Link>
    ),
  },
  {
    accessorKey: 'billing_mode',
    header: 'Mode',
    cell: ({ row }) => (
      <Badge variant='outline'>{row.original.billing_mode}</Badge>
    ),
  },
  {
    id: 'peers',
    header: 'Active / total',
    cell: ({ row }) => (
      <span>
        {row.original.active_peers} / {row.original.total_peers}
      </span>
    ),
  },
  {
    accessorKey: 'credit_rial',
    header: 'Credit (toman)',
    cell: ({ row }) => formatToman(row.original.credit_rial),
  },
  {
    accessorKey: 'debt_rial',
    header: 'Debt (toman)',
    cell: ({ row }) => formatToman(row.original.debt_rial),
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <Badge
        variant={
          row.original.admin_suspended || row.original.billing_suspended
            ? 'destructive'
            : 'outline'
        }
      >
        {row.original.admin_suspended
          ? 'Admin suspended'
          : row.original.billing_suspended
            ? 'Billing suspended'
            : 'Active'}
      </Badge>
    ),
  },
  {
    accessorKey: 'next_invoice_at',
    header: 'Next invoice',
    cell: ({ row }) => dateLabel(row.original.next_invoice_at),
  },
]
