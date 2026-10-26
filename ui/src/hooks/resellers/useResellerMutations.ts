import { useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  ResellerInput,
  OwnershipInput,
  ResellerPeerInput,
  PeerPurchaseInput,
  PeerMutationInput,
  RouterMigrationInput,
  BillingActionInput,
} from '@/schema/billing'
import { toast } from 'sonner'
import {
  submitReceipt,
  reviewReceipt,
  postBillingAction,
  updatePaymentInstructions,
} from '@/api/billing'
import {
  saveReseller,
  setResellerSuspension,
  transferPeerOwnership,
  createResellerPeer,
  purchasePeerOperation,
  mutateResellerPeer,
  beginRouterMigration,
  cancelRouterMigration,
} from '@/api/resellers'

function useResellerMutation<T>(
  action: (input: T) => Promise<unknown>,
  message: string
) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: action,
    onSuccess: () => {
      for (const key of [
        'resellers',
        'reseller',
        'billing',
        'receipts',
        'payment_instructions',
        'peer_purchases',
        'router_migrations',
        'my_peers',
        'peers_list',
      ])
        client.invalidateQueries({ queryKey: [key] })
      toast.success(message)
    },
  })
}

export const useSaveResellerMutation = (id?: number) =>
  useResellerMutation(
    (input: ResellerInput) => saveReseller(input, id),
    id ? 'Reseller updated' : 'Reseller created'
  )

export const useResellerSuspensionMutation = (id: number) =>
  useResellerMutation(
    (input: { suspended: boolean; reason: string }) =>
      setResellerSuspension(id, input),
    'Administrative restriction updated'
  )

export const useTransferPeerMutation = (peerID: number) =>
  useResellerMutation(
    (input: OwnershipInput) => transferPeerOwnership(peerID, input),
    'Ownership updated'
  )

export const useCreateResellerPeerMutation = (id?: number) =>
  useResellerMutation(
    ({ data, key }: { data: ResellerPeerInput; key: string }) =>
      createResellerPeer(data, key, id),
    'Configuration queued for provisioning'
  )

export const usePeerPurchaseMutation = (peerID: number, id?: number) =>
  useResellerMutation(
    ({ data, key }: { data: PeerPurchaseInput; key: string }) =>
      purchasePeerOperation(peerID, data, key, id),
    'Purchase queued for verification'
  )

export const usePeerMaintenanceMutation = (peerID: number, id?: number) =>
  useResellerMutation(
    ({ data, key }: { data: PeerMutationInput; key: string }) =>
      mutateResellerPeer(peerID, data, key, id),
    'Change queued for verification'
  )

export const useBeginRouterMigrationMutation = (id: number) =>
  useResellerMutation(
    ({ data, key }: { data: RouterMigrationInput; key: string }) =>
      beginRouterMigration(id, data, key),
    'Migration queued after successful preflight'
  )

export const useCancelRouterMigrationMutation = (id: number) =>
  useResellerMutation(
    (migrationID: string) => cancelRouterMigration(id, migrationID),
    'Safe cancellation queued'
  )

export const useSubmitReceiptMutation = () =>
  useResellerMutation(
    ({ body, key }: { body: FormData; key: string }) =>
      submitReceipt(body, key),
    'Receipt submitted for review'
  )

export const useReviewReceiptMutation = () =>
  useResellerMutation(reviewReceipt, 'Receipt review saved')

export const useBillingActionMutation = (id: number) =>
  useResellerMutation(
    (input: BillingActionInput) => postBillingAction(id, input),
    'Account action saved'
  )

export const usePaymentInstructionsMutation = () =>
  useResellerMutation(updatePaymentInstructions, 'Payment settings updated')
