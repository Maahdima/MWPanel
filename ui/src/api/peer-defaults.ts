import {
  PeerFormDefaults,
  PeerFormDefaultsResponseSchema,
} from '@/schema/peer-defaults'
import axiosInstance from './axios-instance'

export const fetchPeerFormDefaults = async (): Promise<PeerFormDefaults> => {
  const { data } = await axiosInstance.get('/peer/defaults')
  return PeerFormDefaultsResponseSchema.parse(data).data
}
