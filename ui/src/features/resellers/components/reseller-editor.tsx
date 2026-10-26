import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Reseller,
  ResellerInput,
  ResellerFormSchema,
  ResellerFormValues,
} from '@/schema/billing'
import { toast } from 'sonner'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { useSaveResellerMutation } from '@/hooks/resellers/useResellerMutations'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'
import { SelectDropdown } from '@/components/select-dropdown'
import { ServerName } from '@/components/server-select'
import { bytesToGiB, gibToBytes, inputToman, tomanToRial } from '../money'

function integer(value: string): string | null {
  if (!value.trim()) return null
  if (!/^\d+$/.test(value) || BigInt(value) > 9223372036854775807n)
    throw new Error('Enter a whole nonnegative number within range.')
  return BigInt(value).toString()
}

const financialFields = [
  ['rate', 'Rate override (toman / GiB / 30 days)'],
  ['interval', 'Period (days)'],
  ['grace', 'Grace (days)'],
  ['max_active', 'Maximum active peers'],
  ['max_duration', 'Maximum validity (days)'],
  ['max_peer', 'Maximum quota per peer (GiB)'],
  ['max_total', 'Maximum total configured quota (GiB)'],
  ['fee', 'Recurring fee per period (toman)'],
] as const

export function ResellerEditor({
  current,
  onClose,
}: {
  current?: Reseller
  onClose: () => void
}) {
  const interfaces = useInterfacesListQuery()
  const mutation = useSaveResellerMutation(current?.id)
  const form = useForm<ResellerFormValues>({
    resolver: zodResolver(ResellerFormSchema),
    defaultValues: {
      name: current?.name ?? '',
      username: current?.username ?? '',
      password: '',
      mode: current?.billing_mode ?? 'prepaid',
      rate: current?.rate_rial ? inputToman(current.rate_rial) : '',
      interval: String(current?.interval_days ?? 30),
      grace: String(current?.grace_days ?? 5),
      fee: inputToman(current?.recurring_fee_rial ?? '0'),
      max_active: current?.max_active_peers ?? '',
      max_duration: current?.max_duration_days ?? '',
      max_peer: bytesToGiB(current?.max_peer_traffic_bytes ?? null),
      max_total: bytesToGiB(current?.max_total_traffic_bytes ?? null),
      interface_ids:
        current?.resources.map((resource) => resource.interface_id) ?? [],
      default_interface: current?.default_interface_id?.toString() ?? 'none',
    },
  })
  const selected = useWatch({ control: form.control, name: 'interface_ids' })
  async function submit(values: ResellerFormValues) {
    const value = (key: keyof ResellerFormValues) =>
      String(values[key] ?? '').trim()
    try {
      const input: ResellerInput = {
        name: value('name'),
        billing_mode: values.mode,
        rate_rial: value('rate') ? tomanToRial(value('rate')) : null,
        interval_days: integer(value('interval')) ?? '30',
        grace_days: integer(value('grace')) ?? '5',
        recurring_fee_rial: tomanToRial(value('fee') || '0'),
        interface_ids: values.interface_ids,
        default_interface_id:
          values.default_interface === 'none'
            ? null
            : Number(values.default_interface),
        max_active_peers: integer(value('max_active')),
        max_duration_days: integer(value('max_duration')),
        max_peer_traffic_bytes: gibToBytes(value('max_peer')),
        max_total_traffic_bytes: gibToBytes(value('max_total')),
      }
      if (!current) {
        input.username = value('username')
        input.password = values.password
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
    <Form {...form}>
      <form onSubmit={form.handleSubmit(submit)} className='space-y-4'>
        <div className='grid gap-3 sm:grid-cols-2'>
          {(['name', 'username'] as const).map((name) => (
            <FormField
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <FormItem className='space-y-1'>
                  <FormLabel>{name === 'name' ? 'Name' : 'Username'}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      required
                      maxLength={name === 'name' ? 255 : 64}
                      disabled={name === 'username' && !!current}
                      autoComplete={name === 'username' ? 'off' : undefined}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
          {!current && (
            <FormField
              control={form.control}
              name='password'
              render={({ field }) => (
                <FormItem className='space-y-1 sm:col-span-2'>
                  <FormLabel>Initial password</FormLabel>
                  <FormControl>
                    <PasswordInput
                      {...field}
                      required
                      minLength={12}
                      maxLength={72}
                      autoComplete='new-password'
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}
          <FormField
            control={form.control}
            name='mode'
            render={({ field }) => (
              <FormItem className='space-y-1'>
                <FormLabel>Billing mode</FormLabel>
                <SelectDropdown
                  className='w-full'
                  isControlled
                  defaultValue={field.value}
                  onValueChange={field.onChange}
                  items={[
                    { label: 'Prepaid', value: 'prepaid' },
                    { label: 'Postpaid', value: 'postpaid' },
                  ]}
                />
                <FormMessage />
              </FormItem>
            )}
          />
          {financialFields.map(([name, label]) => (
            <FormField
              key={name}
              control={form.control}
              name={name}
              render={({ field }) => (
                <FormItem className='space-y-1'>
                  <FormLabel>{label}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      inputMode='decimal'
                      placeholder={
                        name === 'rate'
                          ? 'Use global price'
                          : name.startsWith('max')
                            ? 'No limit'
                            : undefined
                      }
                      required={['interval', 'grace', 'fee'].includes(name)}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
        </div>
        <FormField
          control={form.control}
          name='interface_ids'
          render={({ field }) => (
            <FormItem>
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
                    <label
                      key={iface.id}
                      className='flex items-center gap-2 text-sm'
                    >
                      <Checkbox
                        checked={field.value.includes(iface.id)}
                        onCheckedChange={(checked) => {
                          field.onChange(
                            checked
                              ? [...field.value, iface.id]
                              : field.value.filter((id) => id !== iface.id)
                          )
                          if (
                            !checked &&
                            form.getValues('default_interface') ===
                              String(iface.id)
                          )
                            form.setValue('default_interface', 'none')
                        }}
                      />
                      <span>
                        {iface.name} · <ServerName id={iface.server_id} />
                      </span>
                    </label>
                  ))
                )}
              </fieldset>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='default_interface'
          render={({ field }) => (
            <FormItem className='space-y-1'>
              <FormLabel>Default interface</FormLabel>
              <SelectDropdown
                className='w-full'
                isControlled
                defaultValue={field.value}
                onValueChange={field.onChange}
                items={[
                  { label: 'No default', value: 'none' },
                  ...(interfaces.data ?? [])
                    .filter((iface) => selected.includes(iface.id))
                    .map((iface) => ({
                      label: `${iface.name} · Server ${iface.server_id}`,
                      value: String(iface.id),
                    })),
                ]}
              />
              <FormMessage />
            </FormItem>
          )}
        />
        <Button disabled={mutation.isPending || interfaces.isError}>
          {mutation.isPending ? 'Saving...' : 'Save reseller'}
        </Button>
      </form>
    </Form>
  )
}
