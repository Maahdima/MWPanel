import { BillingBook, Reseller } from '@/schema/billing'
import { Peer } from '@/schema/peers'
import {
  resellerAccountLabel,
  peerProvisioningState,
} from '@/lib/reseller-status'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { ServerName } from '@/components/server-select'
import { dateLabel, formatToman, bytesToGiB } from '../money'

export function ResellerOverview({
  reseller,
  book,
  peers,
  onEdit,
}: {
  reseller: Reseller
  book?: BillingBook
  peers?: Peer[]
  onEdit: () => void
}) {
  const pending =
    peers?.filter((peer) =>
      ['pending', 'running'].includes(peerProvisioningState(peer))
    ).length ?? reseller.provisioning_peers
  const failed =
    peers?.filter(
      (peer) =>
        peerProvisioningState(peer) === 'error' ||
        (peer.purchase_pending && peer.operation_error)
    ).length ?? reseller.failed_operations
  const syncErrors =
    peers?.filter((peer) => peer.sync_status === 'error').length ??
    reseller.sync_error_peers
  const syncPending =
    peers?.filter((peer) => peer.sync_status === 'pending').length ??
    reseller.sync_pending_peers
  const counts = [
    ['Total configurations', peers?.length ?? reseller.total_peers],
    [
      'Active',
      peers?.filter((peer) => !peer.disabled).length ?? reseller.active_peers,
    ],
    [
      'Disabled',
      peers?.filter((peer) => peer.disabled).length ?? reseller.disabled_peers,
    ],
    [
      'Expired',
      peers?.filter((peer) => peer.status.includes('expired')).length ??
        reseller.expired_peers,
    ],
  ] as const
  return (
    <div className='space-y-4'>
      <div className='grid grid-cols-2 gap-4 xl:grid-cols-4'>
        {[
          [
            'Account',
            <Badge
              key='account'
              variant={
                reseller.admin_suspended || !reseller.is_active
                  ? 'destructive'
                  : 'outline'
              }
            >
              {resellerAccountLabel(
                reseller.is_active,
                reseller.admin_suspended
              )}
            </Badge>,
          ],
          [
            'Billing',
            <div key='billing' className='flex flex-wrap gap-2'>
              <Badge variant='outline'>{reseller.billing_mode}</Badge>
              <Badge
                variant={
                  book?.state.billing_suspended ? 'destructive' : 'outline'
                }
              >
                {book ? book.state.status.replace(/_/g, ' ') : 'Loading…'}
              </Badge>
            </div>,
          ],
          [
            'Provisioning',
            <p
              key='provisioning'
              className={failed ? 'text-destructive text-sm' : 'text-sm'}
            >
              {failed
                ? `${failed} failed · ${pending} pending`
                : pending
                  ? `${pending} in progress`
                  : reseller.resources.length
                    ? 'Ready to provision'
                    : 'Assign resources to begin'}
            </p>,
          ],
          [
            'Router synchronization',
            <p
              key='sync'
              className={syncErrors ? 'text-destructive text-sm' : 'text-sm'}
            >
              {syncErrors
                ? `${syncErrors} failed · ${syncPending} pending`
                : syncPending
                  ? `${syncPending} pending`
                  : reseller.total_peers
                    ? 'Up to date'
                    : 'No configurations yet'}
            </p>,
          ],
        ].map(([label, value]) => (
          <Card key={String(label)} className='gap-2 py-4'>
            <CardHeader className='px-4'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                {label}
              </CardTitle>
            </CardHeader>
            <CardContent className='px-4'>{value}</CardContent>
          </Card>
        ))}
      </div>
      <div className='grid gap-4 sm:grid-cols-2'>
        <Card>
          <CardHeader>
            <CardTitle className='text-base'>Wallet and obligations</CardTitle>
          </CardHeader>
          <CardContent className='space-y-3 text-sm'>
            <dl className='grid grid-cols-2 gap-x-4 gap-y-2'>
              <dt className='text-muted-foreground'>Available credit</dt>
              <dd className='text-right font-medium tabular-nums'>
                {formatToman(
                  book?.state.available_rial ?? reseller.credit_rial
                )}{' '}
                toman
              </dd>
              <dt className='text-muted-foreground'>Outstanding debt</dt>
              <dd className='text-right tabular-nums'>
                {formatToman(reseller.debt_rial)} toman
              </dd>
              <dt className='text-muted-foreground'>Overdue amount</dt>
              <dd className='text-right tabular-nums'>
                {formatToman(book?.state.overdue_rial ?? '0')} toman
              </dd>
              <dt className='text-muted-foreground'>Reserved credit</dt>
              <dd className='text-right tabular-nums'>
                {formatToman(reseller.reserved_rial)} toman
              </dd>
            </dl>
            {reseller.billing_mode === 'prepaid' &&
              book?.state.available_rial === '0' && (
                <p className='text-muted-foreground'>
                  Top up before a new purchase. Existing paid configurations
                  keep their independent restrictions.
                </p>
              )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className='text-base'>Billing schedule</CardTitle>
          </CardHeader>
          <CardContent className='text-sm'>
            <dl className='grid grid-cols-2 gap-x-4 gap-y-2'>
              <dt className='text-muted-foreground'>Billing period</dt>
              <dd className='text-right'>{reseller.interval_days} days</dd>
              <dt className='text-muted-foreground'>Next invoice</dt>
              <dd className='text-right'>
                {dateLabel(reseller.next_invoice_at)}
              </dd>
              <dt className='text-muted-foreground'>Current due date</dt>
              <dd className='text-right'>
                {book?.state.due_at
                  ? dateLabel(book.state.due_at)
                  : 'Nothing due'}
              </dd>
              <dt className='text-muted-foreground'>Grace period</dt>
              <dd className='text-right'>{reseller.grace_days} days</dd>
              <dt className='text-muted-foreground'>Grace deadline</dt>
              <dd className='text-right'>
                {book?.state.grace_until
                  ? dateLabel(book.state.grace_until)
                  : 'Not in grace'}
              </dd>
            </dl>
          </CardContent>
        </Card>
      </div>
      <div className='grid grid-cols-2 gap-4 lg:grid-cols-4'>
        {counts.map(([label, count]) => (
          <Card key={label} className='gap-1 py-4'>
            <CardHeader className='px-4'>
              <CardTitle className='text-muted-foreground text-sm font-medium'>
                {label}
              </CardTitle>
            </CardHeader>
            <CardContent className='px-4 text-2xl font-semibold tabular-nums'>
              {count}
            </CardContent>
          </Card>
        ))}
      </div>
      <Card>
        <CardHeader className='flex flex-row items-center justify-between gap-3'>
          <CardTitle className='text-base'>Assigned resources</CardTitle>
          <Button variant='outline' size='sm' onClick={onEdit}>
            Edit resources
          </Button>
        </CardHeader>
        <CardContent>
          {reseller.resources.length ? (
            <div className='overflow-x-auto'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Interface</TableHead>
                    <TableHead>Router</TableHead>
                    <TableHead>IP pool</TableHead>
                    <TableHead>Default</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {reseller.resources.map((resource) => (
                    <TableRow key={resource.interface_id}>
                      <TableCell className='font-medium'>
                        {resource.name}
                      </TableCell>
                      <TableCell>
                        <ServerName id={resource.server_id} />
                      </TableCell>
                      <TableCell>
                        {resource.pool_id
                          ? `Pool #${resource.pool_id}`
                          : 'Not assigned'}
                      </TableCell>
                      <TableCell>
                        {reseller.default_interface_id === resource.interface_id
                          ? 'Yes'
                          : '—'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <p className='text-muted-foreground text-sm'>
              Assign a router interface and IP pool before creating
              configurations.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export function ResellerLimits({
  reseller,
  rate,
  onEdit,
}: {
  reseller: Reseller
  rate?: string
  onEdit: () => void
}) {
  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between gap-3'>
        <CardTitle className='text-base'>Pricing and limits</CardTitle>
        <Button variant='outline' onClick={onEdit}>
          Edit settings
        </Button>
      </CardHeader>
      <CardContent className='grid gap-6 text-sm sm:grid-cols-2'>
        <dl className='grid grid-cols-2 gap-x-4 gap-y-3'>
          <dt className='text-muted-foreground'>Rate per GiB / 30 days</dt>
          <dd className='text-right'>
            {rate ? `${formatToman(rate)} toman` : 'Loading…'}
          </dd>
          <dt className='text-muted-foreground'>Price source</dt>
          <dd className='text-right'>
            {reseller.rate_rial === null
              ? 'Global default'
              : 'Reseller override'}
          </dd>
          <dt className='text-muted-foreground'>Recurring fee</dt>
          <dd className='text-right'>
            {formatToman(reseller.recurring_fee_rial)} toman
          </dd>
        </dl>
        <dl className='grid grid-cols-2 gap-x-4 gap-y-3'>
          <dt className='text-muted-foreground'>Maximum active peers</dt>
          <dd className='text-right'>
            {reseller.max_active_peers ?? 'No limit'}
          </dd>
          <dt className='text-muted-foreground'>Maximum validity</dt>
          <dd className='text-right'>
            {reseller.max_duration_days
              ? `${reseller.max_duration_days} days`
              : 'No limit'}
          </dd>
          <dt className='text-muted-foreground'>Per-peer quota</dt>
          <dd className='text-right'>
            {reseller.max_peer_traffic_bytes
              ? `${bytesToGiB(reseller.max_peer_traffic_bytes)} GiB`
              : 'No limit'}
          </dd>
          <dt className='text-muted-foreground'>Total configured quota</dt>
          <dd className='text-right'>
            {reseller.max_total_traffic_bytes
              ? `${bytesToGiB(reseller.max_total_traffic_bytes)} GiB`
              : 'No limit'}
          </dd>
        </dl>
      </CardContent>
    </Card>
  )
}
