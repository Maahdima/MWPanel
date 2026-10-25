import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { ColumnDef } from '@tanstack/react-table'
import { Reseller } from '@/schema/billing'
import { useResellersQuery } from '@/hooks/resellers/useResellerQueries'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { DataTable } from '@/features/shared-components/table/data-table'
import { PaymentSettings } from './billing-panel'
import { dateLabel, formatToman } from './money'
import { PageFrame } from './page-frame'
import { ResellerEditor } from './reseller-editor'

const columns: ColumnDef<Reseller>[] = [
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
export default function ResellersPage() {
  const query = useResellersQuery(),
    [dialog, setDialog] = useState<'create' | 'payment' | null>(null)
  return (
    <PageFrame
      title='Resellers'
      description='Manage accounts, assigned resources and billing.'
      actions={
        <div className='flex gap-2'>
          <Button variant='outline' onClick={() => setDialog('payment')}>
            Payment settings
          </Button>
          <Button onClick={() => setDialog('create')}>Create reseller</Button>
        </div>
      }
    >
      {query.isPending ? (
        <p role='status'>Loading resellers...</p>
      ) : query.isError ? (
        <p role='alert'>
          Unable to load resellers.{' '}
          <Button variant='outline' onClick={() => query.refetch()}>
            Retry
          </Button>
        </p>
      ) : (
        <DataTable data={query.data ?? []} columns={columns} type='Resellers' />
      )}
      <Dialog
        open={dialog !== null}
        onOpenChange={(open) => {
          if (!open) setDialog(null)
        }}
      >
        <DialogContent className='sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>
              {dialog === 'create' ? 'Create reseller' : 'Payment settings'}
            </DialogTitle>
          </DialogHeader>
          <DialogBody>
            {dialog === 'create' ? (
              <ResellerEditor onClose={() => setDialog(null)} />
            ) : dialog === 'payment' ? (
              <PaymentSettings onClose={() => setDialog(null)} />
            ) : null}
          </DialogBody>
        </DialogContent>
      </Dialog>
    </PageFrame>
  )
}
