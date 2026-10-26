import { FormEvent, useRef, useState } from 'react'
import { z } from 'zod'
import { useQuery } from '@tanstack/react-query'
import { createApiResponseSchema } from '@/schema/api-response'
import { Reseller } from '@/schema/billing'
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

const migrationSchema = z.array(
  z.object({
    id: z.string(),
    state: z.string(),
    destination_interface_id: z.number(),
    total_peers: z.number(),
    completed_peers: z.number(),
    attempts: z.number(),
    needs_attention: z.boolean(),
  })
)
export function MigrationButton({ reseller }: { reseller: Reseller }) {
  const [open, setOpen] = useState(false)
  const key = useRef({ signature: '', key: crypto.randomUUID() })
  const history = useQuery({
    queryKey: ['router_migrations', reseller.id],
    queryFn: async () =>
      createApiResponseSchema(migrationSchema).parse(
        (await axiosInstance.get(`/reseller/${reseller.id}/migrations`)).data
      ).data,
    enabled: open || !!reseller.migration_id,
    refetchInterval: open || reseller.migration_id ? 5000 : false,
  })
  const begin = useResellerAction(
    (data: { interface_id: number; reason: string; endpoint: string }) =>
      axiosInstance.post(`/reseller/${reseller.id}/migrations`, data, {
        headers: { 'Idempotency-Key': key.current.key },
      }),
    'Migration queued after successful preflight'
  )
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const form = new FormData(e.currentTarget),
      data = {
        interface_id: Number(form.get('interface')),
        reason: String(form.get('reason')),
        endpoint: String(form.get('endpoint') ?? ''),
      },
      signature = JSON.stringify(data)
    if (signature !== key.current.signature)
      key.current = { signature, key: crypto.randomUUID() }
    await begin.mutateAsync(data)
    await history.refetch()
  }
  const cancel = useResellerAction(
    (id: string) =>
      axiosInstance.post(`/reseller/${reseller.id}/migrations/${id}/cancel`, {
        reason: 'Administrator requested cancellation before routing switch',
      }),
    'Safe cancellation queued'
  )
  const pending =
    history.data?.some((x) => ['pending', 'cancelling'].includes(x.state)) ||
    !!reseller.migration_id
  return (
    <>
      <Button variant='outline' onClick={() => setOpen(true)}>
        {pending ? 'Migration in progress' : 'Migrate router'}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Migrate reseller configurations</DialogTitle>
          </DialogHeader>
          <DialogBody className='space-y-4 pb-5'>
            <form onSubmit={submit} className='space-y-3'>
              <p className='text-muted-foreground text-sm'>
                Existing configurations move to the assigned destination.
                Ownership, billing, usage, expiration and public links are
                preserved. Client addresses may change if needed.
              </p>
              <Label htmlFor='migration-interface'>
                Assigned destination interface
              </Label>
              <select
                id='migration-interface'
                name='interface'
                required
                defaultValue=''
                className='border-input bg-background h-9 w-full rounded-md border px-3'
              >
                <option value='' disabled>
                  Select destination
                </option>
                {reseller.resources.map((resource) => (
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
              <p className='text-muted-foreground text-xs'>
                Assign the destination resource in reseller settings first.
                Source and destination must be on different routers.
              </p>
              <Label htmlFor='migration-endpoint'>
                Public VPN hostname or IP (optional)
              </Label>
              <Input
                id='migration-endpoint'
                name='endpoint'
                maxLength={253}
                placeholder='Defaults to destination server address'
              />
              <Label htmlFor='migration-reason'>Reason</Label>
              <Input
                id='migration-reason'
                name='reason'
                required
                maxLength={1000}
              />
              <p className='text-sm'>
                Clients may need updated downloads. A brief service interruption
                can occur. New configuration changes pause until migration
                finishes.
              </p>
              <Button disabled={begin.isPending || pending}>
                Confirm migration
              </Button>
            </form>
            <div className='max-h-64 overflow-auto border-t pt-3'>
              <p className='mb-2 font-medium'>Recent migrations</p>
              {history.isError ? (
                <p role='alert'>Unable to load migration history.</p>
              ) : history.data?.length ? (
                history.data.map((row) => (
                  <div className='border-b py-2 text-sm' key={row.id}>
                    <div className='flex justify-between'>
                      <span>{row.state}</span>
                      <span>
                        {row.completed_peers} / {row.total_peers} configurations
                      </span>
                    </div>
                    {row.state === 'pending' && row.completed_peers === 0 && (
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={cancel.isPending}
                        onClick={async () => {
                          await cancel.mutateAsync(row.id)
                          await history.refetch()
                        }}
                      >
                        Cancel before switch
                      </Button>
                    )}
                    {row.needs_attention && (
                      <p className='text-muted-foreground'>
                        Verification is pending. Safe retry is automatic;
                        completed steps are retained.
                      </p>
                    )}
                  </div>
                ))
              ) : (
                <p className='text-muted-foreground text-sm'>
                  No migrations yet.
                </p>
              )}
            </div>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  )
}
