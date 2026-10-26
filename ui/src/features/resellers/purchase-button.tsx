import { FormEvent, useRef, useState } from 'react'
import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import { usePeerPurchaseMutation } from '@/hooks/resellers/useResellerMutations'
import {
  useBillingBook,
  usePeerPurchasesQuery,
} from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { formatToman, gibToBytes } from './money'

export function PurchaseButton({
  peer,
  resellerID,
}: {
  peer: Peer
  resellerID?: number
}) {
  const [open, setOpen] = useState(false),
    [operation, setOperation] = useState('add_traffic'),
    [quantity, setQuantity] = useState('1')
  const request = useRef({ signature: '', key: crypto.randomUUID() })
  const book = useBillingBook(resellerID, open)
  const history = usePeerPurchasesQuery(peer.id, resellerID, open)
  const purchase = usePeerPurchaseMutation(peer.id, resellerID)
  let estimate: string | null = null
  try {
    if (book.data && peer.traffic_limit_bytes) {
      const bytes =
        operation === 'add_traffic'
          ? BigInt(gibToBytes(quantity) ?? '0')
          : operation === 'reset_usage'
            ? BigInt(peer.total_usage_bytes)
            : BigInt(peer.traffic_limit_bytes)
      estimate = (
        (bytes * BigInt(book.data.rate_rial) + (1n << 29n)) /
        (1n << 30n)
      ).toString()
    }
  } catch {
    /* Validation is shown on submission. */
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    try {
      const data = {
        operation,
        bytes:
          operation === 'add_traffic' ? (gibToBytes(quantity) ?? '0') : '0',
        days: operation === 'extend_expiry' ? quantity : '0',
      }
      const signature = JSON.stringify(data)
      if (signature !== request.current.signature)
        request.current = { signature, key: crypto.randomUUID() }
      await purchase.mutateAsync({ data, key: request.current.key })
      await history.refetch()
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Unable to submit purchase'
      )
    }
  }
  return (
    <>
      <Button size='sm' variant='outline' onClick={() => setOpen(true)}>
        Purchase
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Purchase for {peer.name}</DialogTitle>
          </DialogHeader>
          <DialogBody className='space-y-4 pb-5'>
            <form onSubmit={submit} className='space-y-3'>
              <Label htmlFor={`purchase-${peer.id}`}>Operation</Label>
              <select
                id={`purchase-${peer.id}`}
                value={operation}
                onChange={(e) => {
                  setOperation(e.target.value)
                  setQuantity('1')
                }}
                className='border-input bg-background h-9 w-full rounded-md border px-3'
              >
                <option value='add_traffic'>Add traffic</option>
                <option value='extend_expiry'>Extend validity</option>
                <option value='reset_usage'>Reset consumed usage</option>
                <option
                  value='full_reset'
                  disabled={peer.configured_duration_days === '0'}
                >
                  Full reset
                </option>
              </select>
              {['add_traffic', 'extend_expiry'].includes(operation) && (
                <>
                  <Label htmlFor={`quantity-${peer.id}`}>
                    {operation === 'add_traffic'
                      ? 'Additional GiB'
                      : 'Additional days'}
                  </Label>
                  <Input
                    id={`quantity-${peer.id}`}
                    value={quantity}
                    onChange={(e) => setQuantity(e.target.value)}
                    inputMode='decimal'
                    required
                  />
                </>
              )}
              <p className='text-muted-foreground text-sm'>
                {operation === 'reset_usage'
                  ? 'The charge uses exact consumed bytes sampled at submission. Expiration stays unchanged.'
                  : operation === 'full_reset'
                    ? 'The full configured quota is charged once. Usage resets and the configured duration restarts today.'
                    : operation === 'extend_expiry'
                      ? 'The full quota is charged once, regardless of the number of added days.'
                      : 'Only the additional quota is charged.'}
              </p>
              {estimate !== null && (
                <p>
                  Estimated charge:{' '}
                  <strong>{formatToman(estimate)} toman</strong>
                  {operation === 'reset_usage' &&
                    ' (usage may increase before submission)'}
                </p>
              )}
              {peer.configured_duration_days === '0' && (
                <p className='text-muted-foreground text-xs'>
                  Full reset requires a recorded original plan duration.
                </p>
              )}
              {book.data?.state.billable_blocked && (
                <p role='alert' className='text-destructive'>
                  New purchases are restricted. Review billing first.
                </p>
              )}
              <Button
                disabled={
                  purchase.isPending ||
                  !book.data ||
                  book.data.state.billable_blocked ||
                  peer.purchase_pending ||
                  peer.migration_pending ||
                  history.data?.some((x) => x.state === 'pending')
                }
              >
                Confirm purchase
              </Button>
            </form>
            <div className='max-h-48 overflow-auto border-t pt-3'>
              <p className='mb-2 font-medium'>Recent purchases</p>
              {history.isError ? (
                <p>Unable to load purchases.</p>
              ) : history.data?.length ? (
                history.data.map((x) => (
                  <div
                    key={x.id}
                    className='flex justify-between border-b py-2 text-sm'
                  >
                    <span>
                      {x.state === 'pending'
                        ? 'Pending router verification'
                        : x.state}
                    </span>
                    <span>{formatToman(x.amount_rial)} toman</span>
                  </div>
                ))
              ) : (
                <p className='text-muted-foreground text-sm'>
                  No purchases yet.
                </p>
              )}
            </div>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  )
}
