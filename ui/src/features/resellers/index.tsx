import { useState } from 'react'
import { useResellersQuery } from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { DataTable } from '@/features/shared-components/table/data-table'
import { PaymentSettings } from './components/billing-panel'
import { ResellerEditor } from './components/reseller-editor'
import { resellersColumns } from './components/resellers-columns'
import { PageFrame } from './page-frame'

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
        <DataTable
          data={query.data ?? []}
          columns={resellersColumns}
          type='Resellers'
        />
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
