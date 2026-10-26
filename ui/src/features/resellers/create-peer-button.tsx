import { FormEvent, useRef, useState } from 'react'
import { z } from 'zod'
import { useQuery } from '@tanstack/react-query'
import { createApiResponseSchema } from '@/schema/api-response'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import {
  useBillingBook,
  useResellerAction,
  useResellerQuery,
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

const resourcesSchema = z.array(
  z.object({
    interface_id: z.number(),
    name: z.string(),
    server_id: z.number(),
    pool_id: z.number().nullable(),
  })
)
export function CreatePeerButton({ resellerID }: { resellerID?: number }) {
  const [open, setOpen] = useState(false),
    [quota, setQuota] = useState('10'),
    [days, setDays] = useState('30')
  const request = useRef({ signature: '', key: crypto.randomUUID() })
  const account = useBillingBook(resellerID, open)
  const reseller = useResellerQuery(resellerID ?? 0)
  const own = useQuery({
    queryKey: ['my_resources'],
    queryFn: async () =>
      createApiResponseSchema(resourcesSchema).parse(
        (await axiosInstance.get('/reseller/me/resources')).data
      ).data,
    enabled: open && !resellerID,
  })
  const resources = resellerID ? reseller.data?.resources : own.data
  const create = useResellerAction(
    (data: {
      name: string
      interface_id: number
      quota_bytes: string
      duration_days: string
    }) =>
      axiosInstance.post(
        resellerID ? `/reseller/${resellerID}/peers` : '/reseller/me/peers',
        data,
        { headers: { 'Idempotency-Key': request.current.key } }
      ),
    'Configuration queued for provisioning'
  )
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
      await create.mutateAsync(data)
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
          </DialogHeader>
          <DialogBody className='space-y-4 pb-5'>
            <form onSubmit={submit} className='space-y-3'>
              <Label htmlFor='reseller-peer-name'>Name</Label>
              <Input
                id='reseller-peer-name'
                name='name'
                required
                maxLength={128}
              />
              <Label htmlFor='reseller-peer-interface'>
                Assigned interface
              </Label>
              <select
                id='reseller-peer-interface'
                name='interface'
                required
                className='border-input bg-background h-9 w-full rounded-md border px-3'
                defaultValue={reseller.data?.default_interface_id ?? ''}
              >
                <option value='' disabled>
                  Select assigned interface
                </option>
                {resources?.map((resource) => (
                  <option
                    key={resource.interface_id}
                    value={resource.interface_id}
                    disabled={!resource.pool_id}
                  >
                    {resource.name} · Server #{resource.server_id}
                    {!resource.pool_id ? ' (no IP pool)' : ''}
                  </option>
                ))}
              </select>
              <div className='grid grid-cols-2 gap-3'>
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
              <Button
                disabled={
                  create.isPending ||
                  !account.data ||
                  account.data.state.billable_blocked ||
                  !resources?.some((resource) => resource.pool_id)
                }
              >
                Confirm purchase and create
              </Button>
            </form>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  )
}
