import { FormEvent, useState } from 'react'
import { Reseller, ResellerInput } from '@/schema/billing'
import { toast } from 'sonner'
import axiosInstance from '@/api/axios-instance'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { useResellerAction } from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ServerName } from '@/components/server-select'
import { bytesToGiB, gibToBytes, inputToman, tomanToRial } from './money'

function integer(value: string): string | null {
  if (!value.trim()) return null
  if (!/^\d+$/.test(value) || BigInt(value) > 9223372036854775807n)
    throw new Error('Enter a whole nonnegative number within range.')
  return BigInt(value).toString()
}
export function ResellerEditor({
  current,
  onClose,
}: {
  current?: Reseller
  onClose: () => void
}) {
  const interfaces = useInterfacesListQuery()
  const [selected, setSelected] = useState<number[]>(
    current?.resources.map((resource) => resource.interface_id) ?? []
  )
  const mutation = useResellerAction(
    (data: ResellerInput) =>
      current
        ? axiosInstance.put(`/reseller/${current.id}`, data)
        : axiosInstance.post('/reseller', data),
    current ? 'Reseller updated' : 'Reseller created'
  )
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const fields = new FormData(event.currentTarget)
    const value = (key: string) => String(fields.get(key) ?? '').trim()
    try {
      const input: ResellerInput = {
        name: value('name'),
        billing_mode: value('mode') as 'prepaid' | 'postpaid',
        rate_rial: value('rate') ? tomanToRial(value('rate')) : null,
        interval_days: integer(value('interval')) ?? '30',
        grace_days: integer(value('grace')) ?? '5',
        recurring_fee_rial: tomanToRial(value('fee') || '0'),
        interface_ids: selected,
        default_interface_id: value('default')
          ? Number(value('default'))
          : null,
        max_active_peers: integer(value('max_active')),
        max_duration_days: integer(value('max_duration')),
        max_peer_traffic_bytes: gibToBytes(value('max_peer')),
        max_total_traffic_bytes: gibToBytes(value('max_total')),
      }
      if (!current) {
        input.username = value('username')
        input.password = String(fields.get('password') ?? '')
      }
      await mutation.mutateAsync(input)
      onClose()
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Unable to save reseller'
      )
    }
  }
  return (
    <form onSubmit={submit} className='space-y-4'>
      <div className='grid gap-3 sm:grid-cols-2'>
        <div className='space-y-1'>
          <Label htmlFor='reseller-name'>Name</Label>
          <Input
            id='reseller-name'
            name='name'
            defaultValue={current?.name}
            required
            maxLength={255}
          />
        </div>
        <div className='space-y-1'>
          <Label htmlFor='reseller-username'>Username</Label>
          <Input
            id='reseller-username'
            name='username'
            defaultValue={current?.username}
            required
            disabled={!!current}
            maxLength={64}
            autoComplete='off'
          />
        </div>
        {!current && (
          <div className='space-y-1 sm:col-span-2'>
            <Label htmlFor='reseller-password'>Initial password</Label>
            <Input
              id='reseller-password'
              name='password'
              type='password'
              required
              minLength={12}
              maxLength={72}
              autoComplete='new-password'
            />
          </div>
        )}
        <div className='space-y-1'>
          <Label htmlFor='reseller-mode'>Billing mode</Label>
          <select
            id='reseller-mode'
            name='mode'
            defaultValue={current?.billing_mode ?? 'prepaid'}
            className='border-input bg-background h-9 w-full rounded-md border px-3 text-sm'
          >
            <option value='prepaid'>Prepaid</option>
            <option value='postpaid'>Postpaid</option>
          </select>
        </div>
        <div className='space-y-1'>
          <Label htmlFor='reseller-rate'>
            Rate override (toman / GiB / 30 days)
          </Label>
          <Input
            id='reseller-rate'
            name='rate'
            inputMode='decimal'
            defaultValue={
              current?.rate_rial ? inputToman(current.rate_rial) : ''
            }
            placeholder='Use global price'
          />
        </div>
        {[
          ['interval', 'Period (days)', String(current?.interval_days ?? 30)],
          ['grace', 'Grace (days)', String(current?.grace_days ?? 5)],
          [
            'max_active',
            'Maximum active peers',
            current?.max_active_peers ?? '',
          ],
          [
            'max_duration',
            'Maximum validity (days)',
            current?.max_duration_days ?? '',
          ],
          [
            'max_peer',
            'Maximum quota per peer (GiB)',
            bytesToGiB(current?.max_peer_traffic_bytes ?? null),
          ],
          [
            'max_total',
            'Maximum total configured quota (GiB)',
            bytesToGiB(current?.max_total_traffic_bytes ?? null),
          ],
          [
            'fee',
            'Recurring fee per period (toman)',
            inputToman(current?.recurring_fee_rial ?? '0'),
          ],
        ].map(([name, label, initial]) => (
          <div className='space-y-1' key={name}>
            <Label htmlFor={`reseller-${name}`}>{label}</Label>
            <Input
              id={`reseller-${name}`}
              name={name}
              inputMode='decimal'
              defaultValue={initial}
              placeholder={name.startsWith('max') ? 'No limit' : undefined}
              required={
                name === 'interval' || name === 'grace' || name === 'fee'
              }
            />
          </div>
        ))}
      </div>
      <fieldset className='space-y-2 rounded-md border p-3'>
        <legend className='px-1 text-sm font-medium'>
          Assigned interfaces and pools
        </legend>
        {interfaces.isPending ? (
          <p>Loading resources...</p>
        ) : interfaces.isError ? (
          <p role='alert'>Unable to load interfaces.</p>
        ) : !interfaces.data?.length ? (
          <p className='text-muted-foreground text-sm'>
            Add an interface and pool before assigning resources.
          </p>
        ) : (
          interfaces.data.map((iface) => (
            <label key={iface.id} className='flex items-center gap-2 text-sm'>
              <Checkbox
                checked={selected.includes(iface.id)}
                onCheckedChange={(checked) =>
                  setSelected((previous) =>
                    checked
                      ? [...previous, iface.id]
                      : previous.filter((id) => id !== iface.id)
                  )
                }
              />
              <span>
                {iface.name} · <ServerName id={iface.server_id} />
              </span>
            </label>
          ))
        )}
      </fieldset>
      <div className='space-y-1'>
        <Label htmlFor='reseller-default'>Default interface</Label>
        <select
          id='reseller-default'
          name='default'
          defaultValue={current?.default_interface_id ?? ''}
          className='border-input bg-background h-9 w-full rounded-md border px-3 text-sm'
        >
          <option value=''>No default</option>
          {interfaces.data
            ?.filter((iface) => selected.includes(iface.id))
            .map((iface) => (
              <option key={iface.id} value={iface.id}>
                {iface.name} · Server {iface.server_id}
              </option>
            ))}
        </select>
      </div>
      <Button disabled={mutation.isPending || interfaces.isError}>
        {mutation.isPending ? 'Saving...' : 'Save reseller'}
      </Button>
    </form>
  )
}
