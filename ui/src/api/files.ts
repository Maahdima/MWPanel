import { saveBlob } from '@/lib/download'
import axiosInstance from './axios-instance'

export async function fetchPrivateFile(path: string): Promise<Blob> {
  const { data } = await axiosInstance.get<Blob>(path, { responseType: 'blob' })
  return data
}
export async function downloadPrivate(path: string, name: string) {
  saveBlob(await fetchPrivateFile(path), name)
}
