import { useQuery } from '@tanstack/react-query'
import { fetchServersList } from '@/api/servers.ts'

export const useServersListQuery = () =>
  useQuery({
    queryKey: ['servers_list'],
    queryFn: fetchServersList,
  })
