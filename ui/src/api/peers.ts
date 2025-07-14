import {
  CreatePeerRequest,
  CreatePeerSchema,
  Peer,
  PeerKeys,
  PeerKeysResponseSchema,
  PeerResponseSchema,
  PeersResponseSchema,
  PeerStats,
  PeerStatsResponseSchema,
  UpdatePeerRequest,
} from '@/schema/peers.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchPeersList = async (): Promise<Peer[]> => {
  const { data } = await axiosInstance.get('/peer')
  const parsed = PeersResponseSchema.parse(data)
  return parsed.data || []
}

export const fetchPeerQRCode = async (
  uuid: string | undefined
): Promise<string> => {
  const response = await axiosInstance.get(`/peer/${uuid}/qrcode`, {
    responseType: 'blob',
  })

  return URL.createObjectURL(response.data)
}

export const fetchPeerConfig = async (
  uuid: string | undefined
): Promise<string> => {
  const response = await axiosInstance.get(`/peer/${uuid}/config`, {
    responseType: 'blob',
  })

  return response.data
}

export const fetchPeerDetails = async (
  uuid: string | undefined
): Promise<PeerStats> => {
  const { data } = await axiosInstance.get(`/peer/${uuid}/details`)
  const parsed = PeerStatsResponseSchema.parse(data)
  return parsed.data
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

export const updatePeer = async (peer: UpdatePeerRequest): Promise<Peer> => {
  const { data } = await axiosInstance.put(`/peer/${peer.id}`, peer)
  const parsed = PeerResponseSchema.parse(data)
  return parsed.data
}

export const deletePeer = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/peer/${id}`)
}

export const fetchPeerKeys = async (): Promise<PeerKeys> => {
  const { data } = await axiosInstance.get('/peer/keys')
  const parsed = PeerKeysResponseSchema.parse(data)
  return parsed.data
}

export const resetPeerUsage = async (id: number): Promise<void> => {
  await axiosInstance.patch(`/peer/${id}/reset-usage`)
}
