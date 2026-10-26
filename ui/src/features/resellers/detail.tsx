import { FormEvent, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useResellerSuspensionMutation } from '@/hooks/resellers/useResellerMutations'
import {
  useBillingBook,
  useResellerPeersQuery,
  useResellerQuery,
} from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { peersColumns } from '@/features/peers/components/peers-columns'
import { PeersDialogs } from '@/features/peers/components/peers-dialogs'
import PeersProvider from '@/features/peers/context/peers-context'
import { DataTable } from '@/features/shared-components/table/data-table'
import { BillingPanel } from './components/billing-panel'
import { CreatePeerButton } from './components/create-peer-button'
import { MigrationButton } from './components/migration-button'
import { ResellerEditor } from './components/reseller-editor'
import { PageFrame } from './page-frame'

export default function ResellerDetail({ id }: { id: number }) {
  const query = useResellerQuery(id),
    book = useBillingBook(id)
  const peers = useResellerPeersQuery(id)
  const [dialog, setDialog] = useState<'settings' | 'suspension' | null>(null)
  const suspension = useResellerSuspensionMutation(id)
  async function saveSuspension(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await suspension.mutateAsync({
      reason: String(new FormData(event.currentTarget).get('reason')),
      suspended: !book.data?.state.administrative_suspended,
    })
    setDialog(null)
  }
  if (query.isPending)
    return (
      <PageFrame title='Reseller'>
        <p>Loading account...</p>
      </PageFrame>
    )
  if (query.isError || !query.data)
    return (
      <PageFrame title='Reseller'>
        <p role='alert'>Unable to load reseller.</p>
      </PageFrame>
    )
  const reseller = query.data
  return (
    <PeersProvider>
      <PageFrame
        title={reseller.name}
        description={`Account: ${reseller.username}`}
        actions={
          <div className='flex flex-wrap gap-2'>
            <CreatePeerButton resellerID={id} />
            <MigrationButton reseller={reseller} />
            <Button variant='outline' asChild>
              <Link to='/resellers'>All resellers</Link>
            </Button>
            <Button variant='outline' onClick={() => setDialog('settings')}>
              Edit settings
            </Button>
            <Button
              variant='outline'
              disabled={!book.data}
              onClick={() => setDialog('suspension')}
            >
              {book.data?.state.administrative_suspended
                ? 'Lift admin suspension'
                : 'Suspend reseller'}
            </Button>
          </div>
        }
      >
        <Tabs defaultValue='billing'>
          <TabsList>
            <TabsTrigger value='billing'>Billing and payments</TabsTrigger>
            <TabsTrigger value='peers'>Configurations</TabsTrigger>
          </TabsList>
          <TabsContent value='billing' className='pt-3'>
            <BillingPanel resellerID={id} />
          </TabsContent>
          <TabsContent value='peers' className='pt-3'>
            {peers.isPending ? (
              <p>Loading configurations...</p>
            ) : peers.isError ? (
              <p role='alert'>Unable to load configurations.</p>
            ) : (
              <DataTable data={peers.data ?? []} columns={peersColumns} />
            )}
          </TabsContent>
        </Tabs>
        <PeersDialogs />
        <Dialog
          open={dialog !== null}
          onOpenChange={(open) => {
            if (!open) setDialog(null)
          }}
        >
          <DialogContent className='sm:max-w-2xl'>
            <DialogHeader>
              <DialogTitle>
                {dialog === 'settings'
                  ? 'Reseller settings'
                  : 'Administrative suspension'}
              </DialogTitle>
              <DialogDescription>
                Manage assigned resources, billing and account restrictions.
              </DialogDescription>
            </DialogHeader>
            {dialog === 'settings' ? (
              <ResellerEditor
                current={reseller}
                onClose={() => setDialog(null)}
              />
            ) : dialog === 'suspension' ? (
              <DialogBody>
                <form onSubmit={saveSuspension} className='space-y-3'>
                  <p className='text-muted-foreground text-sm'>
                    Billing, expiration, quota and manual restrictions remain
                    independent.
                  </p>
                  <Label htmlFor='suspension-reason'>Reason</Label>
                  <Input
                    id='suspension-reason'
                    name='reason'
                    required
                    maxLength={1000}
                  />
                  <Button disabled={suspension.isPending}>
                    Confirm change
                  </Button>
                </form>
              </DialogBody>
            ) : null}
          </DialogContent>
        </Dialog>
      </PageFrame>
    </PeersProvider>
  )
}
