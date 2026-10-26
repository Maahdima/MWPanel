import { FormEvent, useId, useState } from 'react'
import { Peer } from '@/schema/peers'
import { useTransferPeerMutation } from '@/hooks/resellers/useResellerMutations'
import { useResellersQuery } from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogDescription,
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
import { PurchaseButton } from './purchase-button'

export function OwnerCell({ peer }: { peer: Peer }) {
  const formID = useId()
  const { data } = useResellersQuery(),
    [open, setOpen] = useState(false)
  const action = useTransferPeerMutation(peer.id)
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    await action.mutateAsync({
      reseller_id:
        form.get('owner') && form.get('owner') !== 'none'
          ? Number(form.get('owner'))
          : null,
      reason: String(form.get('reason')),
      configured_duration_days: String(form.get('duration') ?? '') || null,
    })
    setOpen(false)
  }
  return (
    <>
      {peer.reseller_id && (
        <PurchaseButton peer={peer} resellerID={peer.reseller_id} />
      )}
      <Button size='sm' variant='ghost' onClick={() => setOpen(true)}>
        {peer.reseller_id
          ? (data?.find((x) => x.id === peer.reseller_id)?.name ??
            `Reseller #${peer.reseller_id}`)
          : 'Superadmin'}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Transfer ownership</DialogTitle>
            <DialogDescription>
              Choose the account responsible for this configuration.
            </DialogDescription>
          </DialogHeader>
          <DialogBody className='space-y-4'>
            <form id={formID} onSubmit={submit} className='space-y-4'>
              <p>{peer.name}</p>
              <div className='space-y-2'>
                <Label htmlFor={`owner-${peer.id}`}>Owner</Label>
                <Select
                  name='owner'
                  defaultValue={peer.reseller_id?.toString() ?? 'none'}
                >
                  <SelectTrigger id={`owner-${peer.id}`} className='w-full'>
                    <SelectValue placeholder='Select' />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='none'>Superadmin</SelectItem>
                    {data?.map((x) => (
                      <SelectItem value={String(x.id)} key={x.id}>
                        {x.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-2'>
                <Label htmlFor={`duration-${peer.id}`}>
                  Original plan duration (days)
                </Label>
                <Input
                  id={`duration-${peer.id}`}
                  name='duration'
                  type='number'
                  min='1'
                  max='36500'
                  defaultValue={
                    peer.configured_duration_days === '0'
                      ? ''
                      : peer.configured_duration_days
                  }
                />
              </div>
              <p className='text-muted-foreground text-xs'>
                Record the original duration to permit full resets of legacy
                peers.
              </p>
              <div className='space-y-2'>
                <Label htmlFor={`reason-${peer.id}`}>Reason</Label>
                <Input
                  id={`reason-${peer.id}`}
                  name='reason'
                  required
                  maxLength={1000}
                />
              </div>
            </form>
          </DialogBody>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              disabled={action.isPending}
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type='submit' form={formID} disabled={action.isPending}>
              Transfer
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
