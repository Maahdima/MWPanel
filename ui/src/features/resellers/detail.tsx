import { FormEvent, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { toast } from 'sonner'
import { useResellerSuspensionMutation } from '@/hooks/resellers/useResellerMutations'
import {
  useBillingBook,
  useResellerPeersQuery,
  useResellerQuery,
} from '@/hooks/resellers/useResellerQueries'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { peersColumns } from '@/features/peers/components/peers-columns'
import { PeersDialogs } from '@/features/peers/components/peers-dialogs'
import PeersProvider from '@/features/peers/context/peers-context'
import { DataTable } from '@/features/shared-components/table/data-table'
import { DataTableSkeleton } from '@/features/shared-components/table/data-table-skeleton'
import { BillingPanel } from './components/billing-panel'
import { CreatePeerButton } from './components/create-peer-button'
import { MigrationButton } from './components/migration-button'
import { ResellerEditor } from './components/reseller-editor'
import {
  ResellerOverview,
  ResellerLimits,
} from './components/reseller-overview'
import { PageFrame } from './page-frame'

type DialogKind = 'settings' | 'suspension'
export default function ResellerDetail({ id }: { id: number }) {
  const query = useResellerQuery(id)
  const book = useBillingBook(id)
  const peers = useResellerPeersQuery(id)
  const [dialog, setDialog] = useState<DialogKind | null>(null)
  const [lastDialog, setLastDialog] = useState<DialogKind>('settings')
  const suspension = useResellerSuspensionMutation(id)
  const openDialog = (kind: DialogKind) => {
    setLastDialog(kind)
    setDialog(kind)
  }
  async function saveSuspension(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      await suspension.mutateAsync({
        reason: String(new FormData(event.currentTarget).get('reason')),
        suspended: !book.data?.state.administrative_suspended,
      })
      setDialog(null)
    } catch {
      toast.error('Unable to update the administrative restriction.')
    }
  }
  if (query.isPending)
    return (
      <PageFrame title='Reseller'>
        <div className='space-y-4'>
          <Skeleton className='h-10 w-64' />
          <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>
            {[0, 1, 2, 3].map((i) => (
              <Skeleton key={i} className='h-24' />
            ))}
          </div>
          <DataTableSkeleton columns={5} />
        </div>
      </PageFrame>
    )
  if (query.isError || !query.data)
    return (
      <PageFrame title='Reseller'>
        <div role='alert' className='space-y-3'>
          <p>Unable to load this reseller.</p>
          <Button variant='outline' onClick={() => query.refetch()}>
            Retry
          </Button>
        </div>
      </PageFrame>
    )
  const reseller = query.data
  const shownDialog = dialog ?? lastDialog
  return (
    <PeersProvider>
      <PageFrame
        title={reseller.name}
        description={`Account: ${reseller.username}`}
        actions={
          <div className='flex flex-wrap gap-2'>
            <Button variant='outline' asChild>
              <Link to='/resellers'>All resellers</Link>
            </Button>
            <CreatePeerButton resellerID={id} />
          </div>
        }
      >
        <Tabs defaultValue='overview' className='space-y-4'>
          <TabsList className='max-w-full justify-start overflow-x-auto'>
            <TabsTrigger className='shrink-0' value='overview'>
              Overview
            </TabsTrigger>
            <TabsTrigger className='shrink-0' value='peers'>
              Configurations
            </TabsTrigger>
            <TabsTrigger className='shrink-0' value='billing'>
              Billing and payments
            </TabsTrigger>
            <TabsTrigger className='shrink-0' value='settings'>
              Settings
            </TabsTrigger>
          </TabsList>
          <TabsContent value='overview' className='space-y-4'>
            {book.isError && (
              <p role='alert' className='text-destructive text-sm'>
                Billing status could not be refreshed.{' '}
                <Button variant='link' size='sm' onClick={() => book.refetch()}>
                  Retry
                </Button>
              </p>
            )}
            <ResellerOverview
              reseller={reseller}
              book={book.data}
              peers={peers.data}
              onEdit={() => openDialog('settings')}
            />
          </TabsContent>
          <TabsContent value='peers'>
            {peers.isPending ? (
              <DataTableSkeleton columns={7} />
            ) : peers.isError ? (
              <p role='alert'>
                Unable to load configurations.{' '}
                <Button variant='outline' onClick={() => peers.refetch()}>
                  Retry
                </Button>
              </p>
            ) : (
              <DataTable
                data={peers.data ?? []}
                columns={peersColumns}
                type='Peers'
              />
            )}
          </TabsContent>
          <TabsContent value='billing'>
            <BillingPanel resellerID={id} />
          </TabsContent>
          <TabsContent value='settings' className='space-y-4'>
            <ResellerLimits
              reseller={reseller}
              rate={book.data?.rate_rial}
              onEdit={() => openDialog('settings')}
            />
            <Card>
              <CardHeader>
                <CardTitle className='text-base'>Router assignment</CardTitle>
              </CardHeader>
              <CardContent className='space-y-3'>
                <p className='text-muted-foreground text-sm'>
                  Assign destination resources in settings before migrating
                  existing configurations. Public links and accounting remain
                  attached to the same peers.
                </p>
                <MigrationButton reseller={reseller} />
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className='text-base'>
                  Administrative restriction
                </CardTitle>
              </CardHeader>
              <CardContent className='space-y-3'>
                <p className='text-muted-foreground text-sm'>
                  This restriction is independent of wallet credit, overdue
                  invoices, peer expiration and manual peer settings.
                </p>
                <Button
                  variant='outline'
                  disabled={!book.data}
                  onClick={() => openDialog('suspension')}
                >
                  {book.data?.state.administrative_suspended
                    ? 'Lift admin suspension'
                    : 'Suspend reseller'}
                </Button>
              </CardContent>
            </Card>
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
                {shownDialog === 'settings'
                  ? 'Reseller settings'
                  : 'Administrative suspension'}
              </DialogTitle>
              <DialogDescription>
                Manage assigned resources, billing and account restrictions.
              </DialogDescription>
            </DialogHeader>
            {shownDialog === 'settings' ? (
              <ResellerEditor
                current={reseller}
                onClose={() => setDialog(null)}
              />
            ) : (
              <>
                <DialogBody>
                  <form
                    id={`reseller-suspension-${id}`}
                    onSubmit={saveSuspension}
                    className='space-y-4'
                  >
                    <p className='text-muted-foreground text-sm'>
                      Billing, expiration, quota and manual restrictions remain
                      independent.
                    </p>
                    <div className='space-y-2'>
                      <Label htmlFor='suspension-reason'>Reason</Label>
                      <Input
                        id='suspension-reason'
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
                    onClick={() => setDialog(null)}
                  >
                    Cancel
                  </Button>
                  <Button
                    type='submit'
                    form={`reseller-suspension-${id}`}
                    disabled={suspension.isPending}
                  >
                    Confirm change
                  </Button>
                </DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </PageFrame>
    </PeersProvider>
  )
}
