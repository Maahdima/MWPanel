import { useState } from 'react'
import { VpnOrder } from '@/schema/vpn-order.ts'
import { toast } from 'sonner'
import { handleServerError } from '@/utils/handle-server-error.ts'
import { useAcceptVpnOrderMutation } from '@/hooks/vpn-orders/useAcceptVpnOrderMutation.ts'
import { Button } from '@/components/ui/button.tsx'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog.tsx'

function formatCard(card: string) {
  const digits = card.replace(/\D/g, '')
  if (digits.length < 4) return card || '—'
  return digits.replace(/(.{4})/g, '$1-').replace(/-$/, '')
}

export function OrderAcceptButton({ order }: { order: VpnOrder }) {
  const [open, setOpen] = useState(false)
  const mutation = useAcceptVpnOrderMutation()

  if (order.status !== 'pending') {
    return <span className='text-muted-foreground'>-</span>
  }

  const details = [
    ['Tracking', order.tracking_code || '—'],
    ['Amount', `${order.amount_tomans.toLocaleString('en-US')} toman`],
    ['Payer card', formatCard(order.payer_card)],
    ['Config', order.peer_name || order.config_name || '—'],
    ['Plan', order.plan_name || order.plan_title || '—'],
    ['Telegram', order.username ? `@${order.username}` : order.chat_id || '—'],
  ]

  return (
    <>
      <Button size='sm' variant='outline' onClick={() => setOpen(true)}>
        Accept
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Accept payment</DialogTitle>
            <DialogDescription>
              The config is activated and the buyer gets a confirmation on
              Telegram.
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <dl className='grid grid-cols-1 gap-3 text-sm'>
              {details.map(([label, value]) => (
                <div key={label} className='min-w-0'>
                  <dt className='text-muted-foreground'>{label}</dt>
                  <dd className='font-medium wrap-break-word'>{value}</dd>
                </div>
              ))}
            </dl>
          </DialogBody>
          <DialogFooter>
            <Button
              variant='outline'
              type='button'
              disabled={mutation.isPending}
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button
              disabled={mutation.isPending}
              onClick={() => {
                mutation.mutate(order.id, {
                  onSuccess: () => {
                    toast.success('Payment accepted')
                    setOpen(false)
                  },
                  onError: (error) => handleServerError(error),
                })
              }}
            >
              Accept payment
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
