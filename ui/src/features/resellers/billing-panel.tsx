import { FormEvent, useEffect, useRef, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Receipt } from '@/schema/billing'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import {
  useBillingBook,
  usePaymentInstructions,
  useReceiptHistory,
  useResellerAction,
} from '@/hooks/resellers/useResellerQueries'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { dateLabel, formatToman, inputToman, tomanToRial } from './money'

export function BillingPanel({ resellerID }: { resellerID?: number }) {
  const account = useBillingBook(resellerID),
    receipts = useReceiptHistory(resellerID),
    instructions = usePaymentInstructions()
  const [preview, setPreview] = useState<string | null>(null),
    [rejectID, setRejectID] = useState<number | null>(null)
  const uploadKey = useRef({ signature: '', key: crypto.randomUUID() })
  useEffect(
    () => () => {
      if (preview) URL.revokeObjectURL(preview)
    },
    [preview]
  )
  const upload = useResellerAction(
    ({ body, key }: { body: FormData; key: string }) =>
      axiosInstance.post('/billing/me/receipts', body, {
        headers: { 'Idempotency-Key': key },
      }),
    'Receipt submitted for review'
  )
  const review = useResellerAction(
    ({
      id,
      action,
      reason,
    }: {
      id: number
      action: 'approve' | 'reject'
      reason?: string
    }) => axiosInstance.post(`/billing/receipts/${id}/${action}`, { reason }),
    'Receipt review saved'
  )
  async function submitReceipt(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = event.currentTarget,
      body = new FormData(form)
    try {
      const amount = tomanToRial(String(body.get('amount')))
      if (BigInt(amount) <= 0n) throw new Error('Amount must be positive.')
      const image = body.get('image')
      if (
        !(image instanceof File) ||
        !image.size ||
        image.size > 5 * 1024 * 1024
      )
        throw new Error('Select a PNG or JPEG receipt up to 5 MiB.')
      body.delete('amount')
      body.set('amount_rial', amount)
      body.set('paid_at', new Date(String(body.get('paid_at'))).toISOString())
      const signature = JSON.stringify([
        amount,
        body.get('paid_at'),
        body.get('transaction_reference'),
        image.name,
        image.size,
        image.lastModified,
      ])
      if (uploadKey.current.signature !== signature)
        uploadKey.current = { signature, key: crypto.randomUUID() }
      await upload.mutateAsync({ body, key: uploadKey.current.key })
      form.reset()
      uploadKey.current = { signature: '', key: crypto.randomUUID() }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Unable to submit receipt'
      )
    }
  }
  async function viewReceipt(receipt: Receipt) {
    try {
      const path = resellerID
        ? `/billing/receipts/${receipt.id}/image`
        : `/billing/me/receipts/${receipt.id}/image`
      const { data } = await axiosInstance.get(path, { responseType: 'blob' })
      setPreview(URL.createObjectURL(data))
    } catch {
      toast.error('Unable to load receipt image')
    }
  }
  if (account.isPending) return <p role='status'>Loading account...</p>
  if (account.isError || !account.data)
    return (
      <p role='alert' className='text-destructive'>
        Unable to load account.{' '}
        <Button variant='outline' onClick={() => account.refetch()}>
          Retry
        </Button>
      </p>
    )
  const book = account.data,
    state = book.state
  return (
    <div className='space-y-5'>
      {(state.billable_blocked ||
        state.billing_suspended ||
        state.administrative_suspended) && (
        <Card className='border-destructive/50'>
          <CardHeader>
            <CardTitle>
              {state.administrative_suspended
                ? 'Administrative suspension'
                : state.billing_suspended
                  ? 'Billing suspension'
                  : 'Payment due'}
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-2'>
            <p>
              {state.administrative_suspended
                ? 'Contact your administrator. Receipt submission remains available.'
                : state.billing_suspended
                  ? 'Submit payment for review to settle overdue obligations.'
                  : 'New purchases are paused while overdue obligations remain unpaid.'}
            </p>
            {state.grace_until > 0 && (
              <p>Grace deadline: {dateLabel(state.grace_until)}</p>
            )}
            {!resellerID && (
              <Button variant='outline' asChild>
                <Link to='/peers'>View configurations</Link>
              </Button>
            )}
          </CardContent>
        </Card>
      )}
      <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
        {[
          ['Balance', state.balance_rial],
          ['Available credit', state.available_rial],
          ['Overdue', state.overdue_rial],
          ['Reserved credit', book.reserved_rial],
        ].map(([label, amount]) => (
          <Card key={label}>
            <CardHeader>
              <CardTitle className='text-muted-foreground text-sm'>
                {label}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className='text-xl font-semibold tabular-nums'>
                {formatToman(amount)}{' '}
                <span className='text-sm font-normal'>toman</span>
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
      <Card>
        <CardContent className='grid gap-2 pt-4 sm:grid-cols-2'>
          <p>
            Mode: <Badge variant='outline'>{book.billing_mode}</Badge>
          </p>
          <p>Next invoice: {dateLabel(state.next_invoice_at)}</p>
          <p>Rate: {formatToman(book.rate_rial)} toman / GiB / 30 days</p>
          <p>
            Period: {book.interval_days} days · Grace: {book.grace_days} days
          </p>
          <p>Recurring fee: {formatToman(book.recurring_fee_rial)} toman</p>
        </CardContent>
      </Card>
      {!resellerID && (
        <div className='grid gap-4 lg:grid-cols-2'>
          <Card>
            <CardHeader>
              <CardTitle>Card-to-card payment</CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              {instructions.isPending ? (
                <p>Loading payment details...</p>
              ) : instructions.isError ? (
                <p role='alert'>Unable to load payment details.</p>
              ) : instructions.data?.card_number ? (
                <>
                  <p className='font-mono text-lg tracking-wider'>
                    {instructions.data.card_number}
                  </p>
                  <p>{instructions.data.cardholder}</p>
                  <p className='text-muted-foreground whitespace-pre-wrap'>
                    {instructions.data.instructions}
                  </p>
                </>
              ) : (
                <p className='text-muted-foreground'>
                  Payment instructions have not been configured. Contact your
                  administrator.
                </p>
              )}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Submit receipt</CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={submitReceipt} className='space-y-3'>
                <div className='space-y-1'>
                  <Label htmlFor='receipt-amount'>Amount (toman)</Label>
                  <Input
                    id='receipt-amount'
                    name='amount'
                    inputMode='decimal'
                    required
                    placeholder='e.g. 100000'
                  />
                </div>
                <div className='space-y-1'>
                  <Label htmlFor='receipt-date'>Payment date and time</Label>
                  <Input
                    id='receipt-date'
                    name='paid_at'
                    type='datetime-local'
                    required
                  />
                </div>
                <div className='space-y-1'>
                  <Label htmlFor='receipt-reference'>
                    Transaction reference
                  </Label>
                  <Input
                    id='receipt-reference'
                    name='transaction_reference'
                    maxLength={128}
                    required
                  />
                </div>
                <div className='space-y-1'>
                  <Label htmlFor='receipt-image'>Receipt image</Label>
                  <Input
                    id='receipt-image'
                    name='image'
                    type='file'
                    accept='image/png,image/jpeg'
                    required
                  />
                  <p className='text-muted-foreground text-xs'>
                    PNG or JPEG, up to 5 MiB. Credit is applied after approval.
                  </p>
                </div>
                <Button type='submit' disabled={upload.isPending}>
                  {upload.isPending ? 'Submitting...' : 'Submit for review'}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      )}
      <Card>
        <CardHeader>
          <CardTitle>Receipt history</CardTitle>
        </CardHeader>
        <CardContent>
          <div className='overflow-x-auto'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Reference</TableHead>
                  <TableHead>Amount (toman)</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Review</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {receipts.isPending ? (
                  <TableRow>
                    <TableCell colSpan={5}>Loading receipts...</TableCell>
                  </TableRow>
                ) : receipts.isError ? (
                  <TableRow>
                    <TableCell colSpan={5}>
                      Unable to load receipt history.
                    </TableCell>
                  </TableRow>
                ) : !receipts.data?.length ? (
                  <TableRow>
                    <TableCell colSpan={5}>No receipts submitted.</TableCell>
                  </TableRow>
                ) : (
                  receipts.data.map((receipt) => (
                    <TableRow key={receipt.id}>
                      <TableCell>{dateLabel(receipt.paid_at)}</TableCell>
                      <TableCell>
                        <Button
                          variant='link'
                          onClick={() => viewReceipt(receipt)}
                        >
                          {receipt.transaction_reference}
                        </Button>
                      </TableCell>
                      <TableCell className='tabular-nums'>
                        {formatToman(receipt.amount_rial)}
                      </TableCell>
                      <TableCell>
                        <Badge variant='outline'>{receipt.state}</Badge>
                        {receipt.review_reason && (
                          <p className='text-muted-foreground mt-1 text-xs'>
                            {receipt.review_reason}
                          </p>
                        )}
                      </TableCell>
                      <TableCell>
                        {resellerID && receipt.state === 'pending' ? (
                          <div className='flex gap-2'>
                            <Button
                              size='sm'
                              disabled={review.isPending}
                              onClick={() =>
                                review.mutate({
                                  id: receipt.id,
                                  action: 'approve',
                                })
                              }
                            >
                              Approve
                            </Button>
                            <Button
                              size='sm'
                              variant='outline'
                              disabled={review.isPending}
                              onClick={() => setRejectID(receipt.id)}
                            >
                              Reject
                            </Button>
                          </div>
                        ) : (
                          dateLabel(receipt.reviewed_at ?? 0)
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
      <div className='grid gap-4 xl:grid-cols-2'>
        <Card>
          <CardHeader>
            <CardTitle>Recent invoices</CardTitle>
          </CardHeader>
          <CardContent>
            <div className='overflow-x-auto'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Period end</TableHead>
                    <TableHead>Billed</TableHead>
                    <TableHead>Due</TableHead>
                    <TableHead>Remaining</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {!book.recent_invoices.length ? (
                    <TableRow>
                      <TableCell colSpan={4}>No invoices yet.</TableCell>
                    </TableRow>
                  ) : (
                    book.recent_invoices.map((invoice) => (
                      <TableRow key={invoice.id}>
                        <TableCell>{dateLabel(invoice.period_end)}</TableCell>
                        <TableCell>{formatToman(invoice.total_rial)}</TableCell>
                        <TableCell>{formatToman(invoice.due_rial)}</TableCell>
                        <TableCell>
                          {formatToman(invoice.remaining_rial)}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Recent ledger entries (toman)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className='overflow-x-auto'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Change</TableHead>
                    <TableHead>Balance</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {!book.recent_ledger.length ? (
                    <TableRow>
                      <TableCell colSpan={4}>No ledger entries.</TableCell>
                    </TableRow>
                  ) : (
                    book.recent_ledger.map((entry) => (
                      <TableRow key={entry.id}>
                        <TableCell>{dateLabel(entry.created_at)}</TableCell>
                        <TableCell>{entry.kind}</TableCell>
                        <TableCell>{formatToman(entry.delta_rial)}</TableCell>
                        <TableCell>
                          {formatToman(entry.balance_after_rial)}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </div>
      {resellerID && <BillingAdminAction id={resellerID} />}
      <Dialog
        open={!!preview}
        onOpenChange={(open) => {
          if (!open) setPreview(null)
        }}
      >
        <DialogContent className='sm:max-w-3xl'>
          <DialogHeader>
            <DialogTitle>Payment receipt</DialogTitle>
          </DialogHeader>
          <DialogBody>
            {preview && (
              <img
                src={preview}
                alt='Payment receipt'
                className='max-h-[70vh] w-full object-contain'
              />
            )}
          </DialogBody>
        </DialogContent>
      </Dialog>
      <Dialog
        open={rejectID !== null}
        onOpenChange={(open) => {
          if (!open) setRejectID(null)
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject receipt</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <form
              className='space-y-3'
              onSubmit={async (event) => {
                event.preventDefault()
                const reason = String(
                  new FormData(event.currentTarget).get('reason')
                )
                if (rejectID !== null) {
                  await review.mutateAsync({
                    id: rejectID,
                    action: 'reject',
                    reason,
                  })
                  setRejectID(null)
                }
              }}
            >
              <Label htmlFor='reject-reason'>Reason</Label>
              <Textarea
                id='reject-reason'
                name='reason'
                required
                maxLength={1000}
              />
              <Button disabled={review.isPending}>Save rejection</Button>
            </form>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function BillingAdminAction({ id }: { id: number }) {
  const [action, setAction] = useState('adjustments')
  const key = useRef(crypto.randomUUID())
  const mutation = useResellerAction(
    (body: { path: string; data: Record<string, unknown> }) =>
      axiosInstance.post(body.path, body.data, {
        headers: { 'Idempotency-Key': key.current },
      }),
    'Account action saved'
  )
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const form = event.currentTarget,
      data = new FormData(form)
    try {
      const reason = String(data.get('reason'))
      const payload: Record<string, unknown> = { reason }
      if (action === 'adjustments') {
        const amount = tomanToRial(String(data.get('amount')))
        payload.delta_rial =
          data.get('direction') === 'debit' ? `-${amount}` : amount
      } else {
        payload.until = Math.trunc(
          new Date(String(data.get('until'))).getTime() / 1000
        )
      }
      await mutation.mutateAsync({
        path: `/billing/account/${id}/${action}`,
        data: payload,
      })
      key.current = crypto.randomUUID()
      form.reset()
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Unable to save action'
      )
    }
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>Account actions</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={submit} className='grid gap-3 sm:grid-cols-2'>
          <div className='space-y-1'>
            <Label htmlFor='billing-action'>Action</Label>
            <select
              id='billing-action'
              value={action}
              onChange={(e) => {
                setAction(e.target.value)
                key.current = crypto.randomUUID()
              }}
              className='border-input bg-background h-9 w-full rounded-md border px-3 text-sm'
            >
              <option value='adjustments'>Manual adjustment</option>
              <option value='grace'>Extend grace</option>
              <option value='override'>Temporary billing reactivation</option>
            </select>
          </div>
          {action === 'adjustments' ? (
            <>
              <div className='space-y-1'>
                <Label htmlFor='adjust-amount'>Amount (toman)</Label>
                <Input
                  id='adjust-amount'
                  name='amount'
                  inputMode='decimal'
                  required
                  onChange={() => {
                    key.current = crypto.randomUUID()
                  }}
                />
              </div>
              <div className='space-y-1'>
                <Label htmlFor='adjust-direction'>Direction</Label>
                <select
                  id='adjust-direction'
                  name='direction'
                  className='border-input bg-background h-9 w-full rounded-md border px-3 text-sm'
                  onChange={() => {
                    key.current = crypto.randomUUID()
                  }}
                >
                  <option value='credit'>Credit / discount</option>
                  <option value='debit'>Debit / penalty</option>
                </select>
              </div>
            </>
          ) : (
            <div className='space-y-1'>
              <Label htmlFor='action-until'>Until</Label>
              <Input
                id='action-until'
                name='until'
                type='datetime-local'
                required
              />
            </div>
          )}
          <div className='space-y-1'>
            <Label htmlFor='action-reason'>Reason</Label>
            <Input
              id='action-reason'
              name='reason'
              required
              maxLength={1000}
              onChange={() => {
                key.current = crypto.randomUUID()
              }}
            />
          </div>
          <div>
            <Button disabled={mutation.isPending}>Save action</Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

export function PaymentSettings({ onClose }: { onClose: () => void }) {
  const query = usePaymentInstructions()
  const mutation = useResellerAction(
    (data: Record<string, string>) =>
      axiosInstance.put('/billing/instructions', data),
    'Payment settings updated'
  )
  if (!query.data)
    return (
      <p>{query.isError ? 'Unable to load payment settings.' : 'Loading...'}</p>
    )
  const settings = query.data
  return (
    <form
      className='space-y-3'
      onSubmit={async (event) => {
        event.preventDefault()
        const data = new FormData(event.currentTarget)
        try {
          await mutation.mutateAsync({
            rate_rial: tomanToRial(String(data.get('rate'))),
            card_number: String(data.get('card')),
            cardholder: String(data.get('holder')),
            instructions: String(data.get('instructions')),
          })
          onClose()
        } catch (error) {
          toast.error(
            error instanceof Error ? error.message : 'Unable to save settings'
          )
        }
      }}
    >
      <Label htmlFor='global-rate'>Price per GiB / 30 days (toman)</Label>
      <Input
        id='global-rate'
        name='rate'
        defaultValue={inputToman(settings.rate_rial)}
        inputMode='decimal'
        required
      />
      <Label htmlFor='destination-card'>Destination card</Label>
      <Input
        id='destination-card'
        name='card'
        defaultValue={settings.card_number}
        maxLength={16}
      />
      <Label htmlFor='cardholder'>Cardholder</Label>
      <Input
        id='cardholder'
        name='holder'
        defaultValue={settings.cardholder}
        maxLength={255}
      />
      <Label htmlFor='payment-instructions'>Instructions</Label>
      <Textarea
        id='payment-instructions'
        name='instructions'
        defaultValue={settings.instructions}
        maxLength={5000}
      />
      <Button disabled={mutation.isPending}>Save payment settings</Button>
    </form>
  )
}
