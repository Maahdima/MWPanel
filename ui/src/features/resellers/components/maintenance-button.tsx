import { FormEvent, useRef, useState } from 'react'
import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import { usePeerMaintenanceMutation } from '@/hooks/resellers/useResellerMutations'
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
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import { bytesToGiB, gibToBytes } from '../money'

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
  const mutation = usePeerMaintenanceMutation(peer.id, resellerID)
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
      await mutation.mutateAsync({ data, key: request.current.key })
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
              <Select
                value={action}
                onValueChange={(selected) => {
                  setAction(selected)
                  setValue(
                    selected === 'rename'
                      ? peer.name
                      : selected === 'set_address'
                        ? peer.allowed_address
                        : selected === 'reduce_quota'
                          ? bytesToGiB(peer.traffic_limit_bytes)
                          : ''
                  )
                }}
              >
                <SelectTrigger id={`maintenance-${peer.id}`} className='w-full'>
                  <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='rename'>Rename</SelectItem>
                  <SelectItem value='set_address'>
                    Change client address
                  </SelectItem>
                  <SelectItem value='regenerate_keys'>
                    Regenerate keys
                  </SelectItem>
                  <SelectItem value='enable'>Enable manually</SelectItem>
                  <SelectItem value='disable'>Disable manually</SelectItem>
                  <SelectItem value='share'>Enable public share</SelectItem>
                  <SelectItem value='unshare'>Disable public share</SelectItem>
                  <SelectItem value='reduce_quota'>Reduce quota</SelectItem>
                  <SelectItem value='delete'>Delete configuration</SelectItem>
                </SelectContent>
              </Select>
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
