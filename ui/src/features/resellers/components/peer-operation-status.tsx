import { Peer } from '@/schema/peers'
import { toast } from 'sonner'
import {
  peerOperationLabel,
  peerProvisioningState,
} from '@/lib/reseller-status'
import { useRetryResellerPeerMutation } from '@/hooks/resellers/useResellerMutations'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

export function PeerOperationStatus({ peer }: { peer: Peer }) {
  const label = peerOperationLabel(peer)
  const failed =
    peerProvisioningState(peer) === 'error' || peer.sync_status === 'error'
  const restrictions =
    peer.restriction_reasons?.filter((reason) => reason !== 'provisioning') ??
    []
  return (
    <div className='min-w-40 space-y-1.5'>
      <Badge variant={failed ? 'destructive' : 'outline'}>{label}</Badge>
      <p className='text-muted-foreground text-xs'>
        Router sync: {peer.sync_status ?? 'unknown'}
      </p>
      {!!restrictions.length && (
        <p className='text-muted-foreground text-xs'>
          Restrictions: {restrictions.join(', ')}
        </p>
      )}
      {peer.operation_error && (
        <p role='alert' className='text-destructive max-w-xs text-xs'>
          {peer.operation_error}
        </p>
      )}
      {!!peer.next_attempt_at && (
        <p className='text-muted-foreground text-xs'>
          Automatic retry:{' '}
          {new Date(peer.next_attempt_at * 1000).toLocaleString()}
        </p>
      )}
    </div>
  )
}

export function PeerRetryButton({
  peer,
  resellerID,
}: {
  peer: Peer
  resellerID: number
}) {
  const retry = useRetryResellerPeerMutation(resellerID, peer.id)
  if (
    !peer.purchase_pending &&
    peer.sync_status !== 'error' &&
    peer.sync_status !== 'pending' &&
    peerProvisioningState(peer) !== 'error'
  )
    return null
  return (
    <Button
      variant='outline'
      size='sm'
      disabled={retry.isPending || peer.migration_pending}
      onClick={() =>
        retry
          .mutateAsync(undefined)
          .catch(() =>
            toast.error(
              'Retry could not be queued. A worker may still be processing this configuration.'
            )
          )
      }
    >
      {retry.isPending ? 'Queuing…' : 'Retry verification'}
    </Button>
  )
}
