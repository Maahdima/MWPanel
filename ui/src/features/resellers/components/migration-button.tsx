import { FormEvent, useRef, useState } from 'react'
import { Reseller } from '@/schema/billing'
import {
  useBeginRouterMigrationMutation,
  useCancelRouterMigrationMutation,
} from '@/hooks/resellers/useResellerMutations'
import { useRouterMigrationsQuery } from '@/hooks/resellers/useResellerQueries'
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

export function MigrationButton({ reseller }: { reseller: Reseller }) {
  const [open, setOpen] = useState(false)
  const key = useRef({ signature: '', key: crypto.randomUUID() })
  const history = useRouterMigrationsQuery(
    reseller.id,
    open || !!reseller.migration_id
  )
  const begin = useBeginRouterMigrationMutation(reseller.id)
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
    await begin.mutateAsync({ data, key: key.current.key })
    await history.refetch()
  }
  const cancel = useCancelRouterMigrationMutation(reseller.id)
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
              <Select name='interface' required defaultValue=''>
                <SelectTrigger id='migration-interface' className='w-full'>
                  <SelectValue placeholder='Select destination' />
                </SelectTrigger>
                <SelectContent>
                  {reseller.resources.map((resource) => (
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
