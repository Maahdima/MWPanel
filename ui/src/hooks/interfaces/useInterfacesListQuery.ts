import { useQuery } from '@tanstack/react-query'
import { fetchInterfacesList } from '@/api/interfaces.ts'

export const useInterfacesListQuery = () =>
  useQuery({
    queryKey: ['interfaces_list'],
    queryFn: fetchInterfacesList,
  })
