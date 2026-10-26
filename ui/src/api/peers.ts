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
  PeerSessions,
  PeerSessionsResponseSchema,
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
import { saveBlob } from '@/lib/download'

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

export const fetchPeerSessions = async (
  id: number,
  scope: 'admin' | 'reseller' = 'admin'
): Promise<PeerSessions> => {
  const { data } = await axiosInstance.get(
    scope === 'reseller'
      ? `/reseller/me/peers/${id}/sessions`
      : `/peer/${id}/sessions`
  )
  const parsed = PeerSessionsResponseSchema.parse(data)
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

  saveBlob(
    blob,
    `traffic-report-${new Date().toISOString().split('T')[0]}.xlsx`
  )
}

export const syncPeers = async (): Promise<void> => {
  await axiosInstance.post('/sync/peers')
}

export const fetchSyncPeers = async (
  interfaceName?: string,
  serverID?: number
): Promise<SyncPeerPreview[]> => {
  const { data } = await axiosInstance.get('/sync/peers', {
    params: { interface: interfaceName, server_id: serverID },
  })
  const parsed = SyncPeersResponseSchema.parse(data)
  return parsed.data || []
}

export const syncSelectedPeers = async (
  payload: SyncPeersRequest
): Promise<void> => {
  const validated = SyncPeersRequestSchema.parse(payload)
  const { server_id, ...body } = validated
  await axiosInstance.post('/sync/peers/selected', body, {
    params: { server_id },
  })
}
