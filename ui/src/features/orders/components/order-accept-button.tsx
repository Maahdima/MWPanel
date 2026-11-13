import { useState } from 'react'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { IconCheck, IconX } from '@tabler/icons-react'
import { Order } from '@/schema/order'
import { toast } from 'sonner'
import { handleServerError } from '@/utils/handle-server-error.ts'
import { useAcceptOrderMutation } from '@/hooks/orders/useAcceptOrderMutation.ts'
import { useRejectOrderMutation } from '@/hooks/orders/useRejectOrderMutation.ts'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu.tsx'

function formatCard(card: string) {
  const digits = card.replace(/\D/g, '')
  if (digits.length < 4) return card || '—'
  return digits.replace(/(.{4})/g, '$1-').replace(/-$/, '')
}

function canReview(order: Order) {
  return order.status === 'pending' || order.status === 'expired'
}

export function OrderAcceptButton({ order }: { order: Order }) {
  const [acceptOpen, setAcceptOpen] = useState(false)
  const [rejectOpen, setRejectOpen] = useState(false)
  const accept = useAcceptOrderMutation()
  const reject = useRejectOrderMutation()

  if (!canReview(order)) {
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
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button
            variant='ghost'
            className='data-[state=open]:bg-muted flex h-8 w-8 p-0'
          >
            <DotsHorizontalIcon className='h-4 w-4' />
            <span className='sr-only'>Open menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem onClick={() => setAcceptOpen(true)}>
            Accept
            <DropdownMenuShortcut>
              <IconCheck size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => setRejectOpen(true)}
            className='text-red-500!'
          >
            Reject
            <DropdownMenuShortcut>
              <IconX size={16} />
            </DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <Dialog open={acceptOpen} onOpenChange={setAcceptOpen}>
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
              disabled={accept.isPending}
              onClick={() => setAcceptOpen(false)}
            >
              Cancel
            </Button>
            <Button
              disabled={accept.isPending}
              onClick={() => {
                accept.mutate(order.id, {
                  onSuccess: () => {
                    toast.success('Payment accepted')
                    setAcceptOpen(false)
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
      <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject payment</DialogTitle>
            <DialogDescription>
              The order is closed and the buyer is told the payment was not
              accepted. The config is not changed.
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
              disabled={reject.isPending}
              onClick={() => setRejectOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant='destructive'
              disabled={reject.isPending}
              onClick={() => {
                reject.mutate(order.id, {
                  onSuccess: () => {
                    toast.success('Payment rejected')
                    setRejectOpen(false)
                  },
                  onError: (error) => handleServerError(error),
                })
              }}
            >
              Reject payment
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
