import { FormEvent, useRef, useState } from 'react'
import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import { useResellerAction } from '@/hooks/resellers/useResellerQueries'
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
import { bytesToGiB, gibToBytes } from './money'

export function MaintenanceButton({
  peer,
  resellerID,
}: {
  peer: Peer
  resellerID?: number
}) {
  const [open, setOpen] = useState(false),
    [action, setAction] = useState('rename'),
    [value, setValue] = useState(peer.name)
  const request = useRef({ signature: '', key: crypto.randomUUID() })
  const mutation = useResellerAction(
    (data: { action: string; value: string; enabled: boolean | null }) =>
      axiosInstance.post(
        resellerID
          ? `/reseller/${resellerID}/peers/${peer.id}/mutations`
          : `/reseller/me/peers/${peer.id}/mutations`,
        data,
        { headers: { 'Idempotency-Key': request.current.key } }
      ),
    'Change queued for verification'
  )
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    try {
      const data = {
        action:
          action === 'enable' || action === 'disable'
            ? 'set_manual'
            : action === 'share' || action === 'unshare'
              ? 'set_share'
              : action,
        value: ['rename', 'set_address', 'reduce_quota'].includes(action)
          ? action === 'reduce_quota'
            ? (gibToBytes(value) ?? '0')
            : value
          : '',
        enabled: ['enable', 'share'].includes(action)
          ? true
          : ['disable', 'unshare'].includes(action)
            ? false
            : null,
      }
      const signature = JSON.stringify(data)
      if (signature !== request.current.signature)
        request.current = { signature, key: crypto.randomUUID() }
      await mutation.mutateAsync(data)
      setOpen(false)
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Unable to queue change'
      )
    }
  }
  return (
    <>
      <Button variant='ghost' size='sm' onClick={() => setOpen(true)}>
        Manage
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Manage {peer.name}</DialogTitle>
          </DialogHeader>
          <DialogBody className='space-y-4 pb-5'>
            <form onSubmit={submit} className='space-y-3'>
              <Label htmlFor={`maintenance-${peer.id}`}>Free operation</Label>
              <select
                id={`maintenance-${peer.id}`}
                value={action}
                onChange={(e) => {
                  setAction(e.target.value)
                  setValue(
                    e.target.value === 'rename'
                      ? peer.name
                      : e.target.value === 'set_address'
                        ? peer.allowed_address
                        : e.target.value === 'reduce_quota'
                          ? bytesToGiB(peer.traffic_limit_bytes)
                          : ''
                  )
                }}
                className='border-input bg-background h-9 w-full rounded-md border px-3'
              >
                <option value='rename'>Rename</option>
                <option value='set_address'>Change client address</option>
                <option value='regenerate_keys'>Regenerate keys</option>
                <option value='enable'>Enable manually</option>
                <option value='disable'>Disable manually</option>
                <option value='share'>Enable public share</option>
                <option value='unshare'>Disable public share</option>
                <option value='reduce_quota'>Reduce quota</option>
                <option value='delete'>Delete configuration</option>
              </select>
              {['rename', 'set_address', 'reduce_quota'].includes(action) && (
                <>
                  <Label htmlFor={`maintenance-value-${peer.id}`}>
                    {action === 'rename'
                      ? 'Name'
                      : action === 'set_address'
                        ? 'Client IPv4 address (/32)'
                        : 'New quota (GiB)'}
                  </Label>
                  <Input
                    id={`maintenance-value-${peer.id}`}
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    required
                  />
                </>
              )}
              <p className='text-muted-foreground text-sm'>
                This operation has no charge. Existing billing, expiration and
                quota restrictions still apply.
              </p>
              {action === 'delete' && (
                <p className='text-destructive'>
                  The configuration and its router objects will be removed.
                  Previous charges are retained.
                </p>
              )}
              {action === 'regenerate_keys' && (
                <p className='text-muted-foreground'>
                  Clients will need to download an updated configuration.
                </p>
              )}
              {action === 'reduce_quota' && (
                <p className='text-muted-foreground'>
                  Reducing quota does not refund previous charges.
                </p>
              )}
              <Button
                variant={action === 'delete' ? 'destructive' : 'default'}
                disabled={
                  mutation.isPending ||
                  peer.purchase_pending ||
                  peer.migration_pending
                }
              >
                Confirm change
              </Button>
            </form>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  )
}
