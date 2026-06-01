import {
  CreateInterfaceRequest,
  CreateInterfaceSchema,
  Interface,
  InterfaceResponseSchema,
  InterfacesResponseSchema,
  UpdateInterfaceRequest,
} from '@/schema/interfaces.ts'
import {
  SyncInterfacePreview,
  SyncInterfacesRequest,
  SyncInterfacesRequestSchema,
  SyncInterfacesResponseSchema,
} from '@/schema/sync.ts'
import axiosInstance from '@/api/axios-instance.ts'

export const fetchInterfacesList = async (): Promise<Interface[]> => {
  const { data } = await axiosInstance.get('/interface')
  const parsed = InterfacesResponseSchema.parse(data)
  return parsed.data || []
}

export const createInterface = async (
  wgInterface: CreateInterfaceRequest
): Promise<Interface> => {
  const validated = CreateInterfaceSchema.parse(wgInterface)
  const { data } = await axiosInstance.post('/interface', validated)
  const parsed = InterfaceResponseSchema.parse(data)
  return parsed.data
}

export const updateInterfaceStatus = async (id: number): Promise<void> => {
  await axiosInstance.patch(`/interface/${id}/status`)
}

export const updateInterface = async (
  wgInterface: UpdateInterfaceRequest
): Promise<Interface> => {
  const { data } = await axiosInstance.put(
    `/interface/${wgInterface.id}`,
    wgInterface
  )
  const parsed = InterfaceResponseSchema.parse(data)
  return parsed.data
}

export const deleteInterface = async (id: number): Promise<void> => {
  await axiosInstance.delete(`/interface/${id}`)
}

export const syncInterfaces = async (): Promise<void> => {
  await axiosInstance.post('/sync/interfaces')
}

export const fetchSyncInterfaces = async (): Promise<
  SyncInterfacePreview[]
> => {
  const { data } = await axiosInstance.get('/sync/interfaces')
  const parsed = SyncInterfacesResponseSchema.parse(data)
  return parsed.data || []
}

export const syncSelectedInterfaces = async (
  payload: SyncInterfacesRequest
): Promise<void> => {
  const validated = SyncInterfacesRequestSchema.parse(payload)
  await axiosInstance.post('/sync/interfaces/selected', validated)
}

