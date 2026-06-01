import {
  CreatePeerRequest,
  CreatePeerSchema,
  FetchPeerAllowedAddress,
  Peer,
  PeerAllowedAddress,
  PeerAllowedAddressResponseSchema,
  PeerCredentials,
  PeerCredentialsResponseSchema,
  PeerResponseSchema,
  PeerShare,
  PeerShareResponseSchema,
  PeersResponseSchema,
  UpdatePeerRequest,
  UpdatePeerShareExpireRequest,
} from '@/schema/peers.ts'
import {
  SyncPeerPreview,
  SyncPeersRequest,
  SyncPeersRequestSchema,
  SyncPeersResponseSchema,
} from '@/schema/sync.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchPeersList = async (): Promise<Peer[]> => {
  const { data } = await axiosInstance.get('/peer')
  const parsed = PeersResponseSchema.parse(data)
  return parsed.data || []
}

export const fetchPeerQRCode = async (id: number): Promise<string> => {
  const response = await axiosInstance.get(`/peer/${id}/qrcode`, {
    responseType: 'blob',
  })

  return URL.createObjectURL(response.data)
}

export const fetchPeerConfig = async (id: number): Promise<string> => {
  const response = await axiosInstance.get(`/peer/${id}/config`, {
    responseType: 'blob',
  })

  return response.data
}

export const createPeer = async (peer: CreatePeerRequest): Promise<Peer> => {
  const validated = CreatePeerSchema.parse(peer)
  const { data } = await axiosInstance.post('/peer', validated)
  const parsed = PeerResponseSchema.parse(data)
  return parsed.data
}

export const updatePeerStatus = async (id: number): Promise<void> => {
  await axiosInstance.patch(`/peer/${id}/status`)
}

export const updatePeerShareStatus = async (id: number): Promise<void> => {
  await axiosInstance.patch(`/peer/${id}/share/status`)
}

export const updatePeer = async (peer: UpdatePeerRequest): Promise<Peer> => {
  const { data } = await axiosInstance.put(`/peer/${peer.id}`, peer)
  const parsed = PeerResponseSchema.parse(data)
  return parsed.data
}

export const deletePeer = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/peer/${id}`)
}

export const fetchPeerAllowedAddress = async (
  iface: FetchPeerAllowedAddress
): Promise<PeerAllowedAddress> => {
  const { data } = await axiosInstance.post(`/peer/allowed-address`, iface)
  const parsed = PeerAllowedAddressResponseSchema.parse(data)
  return parsed.data
}

export const fetchPeerCredentials = async (): Promise<PeerCredentials> => {
  const { data } = await axiosInstance.get('/peer/credentials')
  const parsed = PeerCredentialsResponseSchema.parse(data)
  return parsed.data
}

export const fetchPeerShareStatus = async (id: number): Promise<PeerShare> => {
  const { data } = await axiosInstance.get(`/peer/${id}/share`)
  const parsed = PeerShareResponseSchema.parse(data)
  return parsed.data
}

export const updatePeerShareExpire = async (
  peer: UpdatePeerShareExpireRequest
): Promise<void> => {
  await axiosInstance.patch(`/peer/${peer.id}/share/expire`, peer)
}

export const resetPeerUsage = async (id: number): Promise<void> => {
  await axiosInstance.patch(`/peer/${id}/reset-usage`)
}

export const resetPeerUsages = async (): Promise<void> => {
  await axiosInstance.patch(`/peer/reset-usage`)
}

export const exportTrafficExcel = async (): Promise<void> => {
  const response = await axiosInstance.post(`/peer/traffic/export`, null, {
    responseType: 'blob',
  })

  const blob = new Blob([response.data], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })

  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')

  link.href = url
  link.download = `traffic-report-${new Date().toISOString().split('T')[0]}.xlsx`

  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export const syncPeers = async (): Promise<void> => {
  await axiosInstance.post('/sync/peers')
}

export const fetchSyncPeers = async (
  interfaceName?: string
): Promise<SyncPeerPreview[]> => {
  const { data } = await axiosInstance.get('/sync/peers', {
    params: interfaceName ? { interface: interfaceName } : undefined,
  })
  const parsed = SyncPeersResponseSchema.parse(data)
  return parsed.data || []
}

export const syncSelectedPeers = async (
  payload: SyncPeersRequest
): Promise<void> => {
  const validated = SyncPeersRequestSchema.parse(payload)
  await axiosInstance.post('/sync/peers/selected', validated)
}

