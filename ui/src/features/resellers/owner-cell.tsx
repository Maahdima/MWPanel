import { FormEvent, useState } from 'react'
import { Peer } from '@/schema/peers'
import axiosInstance from '@/api/axios-instance'
import {
  useResellersQuery,
  useResellerAction,
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
import { PurchaseButton } from './purchase-button'

export function OwnerCell({ peer }: { peer: Peer }) {
  const { data } = useResellersQuery(),
    [open, setOpen] = useState(false)
  const action = useResellerAction(
    (payload: {
      reseller_id: number | null
      reason: string
      configured_duration_days: string | null
    }) => axiosInstance.put(`/peer/${peer.id}/owner`, payload),
    'Ownership updated'
  )
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    await action.mutateAsync({
      reseller_id: form.get('owner') ? Number(form.get('owner')) : null,
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
          </DialogHeader>
          <DialogBody className='space-y-4 pb-5'>
            <form onSubmit={submit} className='space-y-3'>
              <p>{peer.name}</p>
              <Label htmlFor={`owner-${peer.id}`}>Owner</Label>
              <select
                id={`owner-${peer.id}`}
                name='owner'
                defaultValue={peer.reseller_id ?? ''}
                className='border-input bg-background h-9 w-full rounded-md border px-3'
              >
                <option value=''>Superadmin</option>
                {data?.map((x) => (
                  <option value={x.id} key={x.id}>
                    {x.name}
                  </option>
                ))}
              </select>
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
              <p className='text-muted-foreground text-xs'>
                Record the original duration to permit full resets of legacy
                peers.
              </p>
              <Label htmlFor={`reason-${peer.id}`}>Reason</Label>
              <Input
                id={`reason-${peer.id}`}
                name='reason'
                required
                maxLength={1000}
              />
              <Button disabled={action.isPending}>Transfer</Button>
            </form>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  )
}
