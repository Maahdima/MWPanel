import { useQuery } from '@tanstack/react-query'
import {
  fetchPrincipal,
  fetchBillingBook,
  fetchReceiptHistory,
  fetchPaymentInstructions,
} from '@/api/billing'
import {
  fetchResellers,
  fetchReseller,
  fetchResellerPeers,
  fetchMyResources,
  fetchPeerPurchases,
  fetchRouterMigrations,
} from '@/api/resellers'
import { useAuthStore } from '@/stores/authStore'

export const usePrincipal = () => {
  const token = useAuthStore((state) => state.auth.accessToken)
  return useQuery({
    queryKey: ['principal'],
    queryFn: fetchPrincipal,
    enabled: !!token,
    staleTime: 10000,
    retry: false,
  })
}

export const useResellersQuery = () =>
  useQuery({ queryKey: ['resellers'], queryFn: fetchResellers })

export const useResellerQuery = (id: number) =>
  useQuery({
    queryKey: ['reseller', id],
    queryFn: () => fetchReseller(id),
    enabled: id > 0,
  })

export const useBillingBook = (id?: number, enabled = true) =>
  useQuery({
    queryKey: ['billing', id ?? 'me'],
    queryFn: () => fetchBillingBook(id),
    enabled,
    refetchInterval: 15000,
  })

export const useReceiptHistory = (id?: number) =>
  useQuery({
    queryKey: ['receipts', id ?? 'me'],
    queryFn: () => fetchReceiptHistory(id),
    refetchInterval: 15000,
  })

export const usePaymentInstructions = () =>
  useQuery({
    queryKey: ['payment_instructions'],
    queryFn: fetchPaymentInstructions,
  })

export const useResellerPeersQuery = (id?: number) =>
  useQuery({
    queryKey: ['peers_list', 'reseller', id ?? 'me'],
    queryFn: () => fetchResellerPeers(id),
    refetchInterval: id ? false : 15000,
  })

export const useMyResourcesQuery = (enabled: boolean) =>
  useQuery({ queryKey: ['my_resources'], queryFn: fetchMyResources, enabled })

export const usePeerPurchasesQuery = (
  peerID: number,
  id: number | undefined,
  open: boolean
) =>
  useQuery({
    queryKey: ['peer_purchases', id ?? 'me', peerID],
    queryFn: () => fetchPeerPurchases(peerID, id),
    enabled: open,
    refetchInterval: open ? 5000 : false,
  })

export const useRouterMigrationsQuery = (id: number, enabled: boolean) =>
  useQuery({
    queryKey: ['router_migrations', id],
    queryFn: () => fetchRouterMigrations(id),
    enabled,
    refetchInterval: enabled ? 5000 : false,
  })
