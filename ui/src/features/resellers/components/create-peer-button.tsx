import { FormEvent, useId, useRef, useState } from 'react'
import { toast } from 'sonner'
import { useCreateResellerPeerMutation } from '@/hooks/resellers/useResellerMutations'
import {
  useBillingBook,
  useMyResourcesQuery,
  useResellerQuery,
} from '@/hooks/resellers/useResellerQueries'
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
import { formatToman, gibToBytes } from '../money'

export function CreatePeerButton({ resellerID }: { resellerID?: number }) {
  const formID = useId()
  const [open, setOpen] = useState(false),
    [quota, setQuota] = useState('10'),
    [days, setDays] = useState('30')
  const request = useRef({ signature: '', key: crypto.randomUUID() })
  const account = useBillingBook(resellerID, open)
  const reseller = useResellerQuery(resellerID ?? 0)
  const own = useMyResourcesQuery(open && !resellerID)
  const resources = resellerID ? reseller.data?.resources : own.data
  const create = useCreateResellerPeerMutation(resellerID)
  let estimated: string | null = null
  try {
    if (account.data && /^\d+$/.test(days)) {
      const bytes = BigInt(gibToBytes(quota) ?? '0'),
        duration = BigInt(days),
        denominator = (1n << 30n) * 30n
      estimated = (
        (bytes *
          BigInt(account.data.rate_rial) *
          (duration > 30n ? duration : 30n) +
          denominator / 2n) /
        denominator
      ).toString()
    }
  } catch {
    /* Submission reports invalid input. */
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    try {
      const form = new FormData(e.currentTarget),
        data = {
          name: String(form.get('name')),
          interface_id: Number(form.get('interface')),
          quota_bytes: gibToBytes(quota) ?? '0',
          duration_days: days,
        }
      const signature = JSON.stringify(data)
      if (signature !== request.current.signature)
        request.current = { signature, key: crypto.randomUUID() }
      await create.mutateAsync({ data, key: request.current.key })
      setOpen(false)
    } catch (error) {
      toast.error(
        error instanceof Error
          ? error.message
          : 'Unable to create configuration'
      )
    }
  }
  return (
    <>
      <Button onClick={() => setOpen(true)}>Create configuration</Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create configuration</DialogTitle>
            <DialogDescription>
              Choose an assigned interface, quota and validity.
            </DialogDescription>
          </DialogHeader>
          <DialogBody className='space-y-4'>
            <form id={formID} onSubmit={submit} className='space-y-4'>
              <div className='space-y-2'>
                <Label htmlFor='reseller-peer-name'>Name</Label>
                <Input
                  id='reseller-peer-name'
                  name='name'
                  required
                  maxLength={128}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='reseller-peer-interface'>
                  Assigned interface
                </Label>
                <Select
                  name='interface'
                  required
                  defaultValue={
                    reseller.data?.default_interface_id?.toString() ?? ''
                  }
                >
                  <SelectTrigger
                    id='reseller-peer-interface'
                    className='w-full'
                  >
                    <SelectValue placeholder='Select assigned interface' />
                  </SelectTrigger>
                  <SelectContent>
                    {resources?.map((resource) => (
                      <SelectItem
                        key={resource.interface_id}
                        value={String(resource.interface_id)}
                        disabled={!resource.pool_id}
                      >
                        {resource.name} · Server #{resource.server_id}
                        {!resource.pool_id ? ' (no IP pool)' : ''}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='grid gap-4 sm:grid-cols-2'>
                <div className='space-y-2'>
                  <Label htmlFor='reseller-peer-quota'>Quota (GiB)</Label>
                  <Input
                    id='reseller-peer-quota'
                    required
                    value={quota}
                    onChange={(e) => setQuota(e.target.value)}
                    inputMode='decimal'
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='reseller-peer-days'>Validity (days)</Label>
                  <Input
                    id='reseller-peer-days'
                    type='number'
                    min='1'
                    max='36500'
                    required
                    value={days}
                    onChange={(e) => setDays(e.target.value)}
                  />
                </div>
              </div>
              <p className='text-muted-foreground text-sm'>
                Creation has a 30-day minimum price. Longer plans scale by their
                duration. Credentials and addressing are generated securely on
                the assigned router.
              </p>
              {estimated !== null && (
                <p>
                  Charge at current rate:{' '}
                  <strong>{formatToman(estimated)} toman</strong>
                </p>
              )}
              {account.data?.state.billable_blocked && (
                <p className='text-destructive'>
                  Review billing before creating configurations.
                </p>
              )}
              {resources?.length === 0 && (
                <p className='text-muted-foreground'>
                  No resources assigned. Contact the administrator.
                </p>
              )}
            </form>
          </DialogBody>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              disabled={create.isPending}
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type='submit'
              form={formID}
              disabled={
                create.isPending ||
                !account.data ||
                account.data.state.billable_blocked ||
                !resources?.some((resource) => resource.pool_id)
              }
            >
              Confirm purchase and create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
