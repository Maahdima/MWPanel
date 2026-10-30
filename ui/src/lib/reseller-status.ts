import type { Peer } from '@/schema/peers'

type PeerState = Pick<
  Peer,
  | 'disabled'
  | 'migration_pending'
  | 'purchase_pending'
  | 'manual_disabled'
  | 'restriction_reasons'
  | 'sync_status'
  | 'provisioning_status'
  | 'operation_error'
>

export function peerProvisioningState(peer: PeerState) {
  return (
    peer.provisioning_status ??
    (peer.restriction_reasons?.includes('provisioning') ? 'pending' : 'ready')
  )
}

export function peerOperationLabel(peer: PeerState) {
  if (peer.migration_pending) return 'Migrating'
  const setup = peerProvisioningState(peer)
  if (setup === 'error') return 'Provisioning failed'
  if (setup !== 'ready') return 'Provisioning'
  if (peer.purchase_pending)
    return peer.operation_error ? 'Change failed' : 'Applying change'
  if (peer.restriction_reasons?.includes('expired')) return 'Expired'
  if (peer.restriction_reasons?.includes('quota')) return 'Quota reached'
  if (peer.manual_disabled || peer.restriction_reasons?.includes('manual'))
    return 'Manually disabled'
  if (
    peer.restriction_reasons?.some((reason) =>
      ['billing', 'administrative'].includes(reason)
    )
  )
    return 'Restricted'
  if (peer.disabled && peer.sync_status === 'pending')
    return 'Activation pending'
  if (peer.disabled && peer.sync_status === 'error') return 'Activation failed'
  return peer.disabled ? 'Disabled' : 'Enabled'
}

export function resellerAccountLabel(
  active: boolean,
  administrativeSuspended: boolean
) {
  if (!active) return 'Inactive'
  return administrativeSuspended ? 'Admin suspended' : 'Active'
}
