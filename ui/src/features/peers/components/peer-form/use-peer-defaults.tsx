'use client'

import { useEffect, useRef } from 'react'
import { UseFormReturn } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { useServersListQuery } from '@/hooks/servers/useServersListQuery'

interface UsePeerDefaultsProps {
  isEdit: boolean
  form: UseFormReturn<CreatePeerRequest>
  shouldRefetch: boolean
}

export function usePeerDefaults({
  isEdit,
  form,
  shouldRefetch,
}: UsePeerDefaultsProps) {
  const {
    data: interfacesList = [],
    isLoading: isInterfacesLoading,
    error: interfacesError,
    refetch: refetchInterfaces,
  } = useInterfacesListQuery()

  const {
    data: serversList = [],
    isLoading: isServersLoading,
    error: serversError,
    refetch: refetchServers,
  } = useServersListQuery()

  const hasSetDefaults = useRef(false)

  useEffect(() => {
    if (!isEdit && shouldRefetch) {
      refetchInterfaces()
      refetchServers()
    }
  }, [shouldRefetch, isEdit, refetchInterfaces, refetchServers])

  useEffect(() => {
    if (
      !isEdit &&
      !hasSetDefaults.current &&
      interfacesList.length > 0 &&
      serversList.length > 0 &&
      !isInterfacesLoading &&
      !isServersLoading
    ) {
      const defaultInterface = interfacesList[0]
      const defaultServer = serversList[0]

      form.reset((prev) => ({
        ...prev,
        interface_name: defaultInterface.name,
        interface_id: defaultInterface.id,
        endpoint: defaultServer.ip_address,
        persistent_keepalive: '00:00:25',
      }))

      hasSetDefaults.current = true
    }
  }, [
    isEdit,
    form,
    interfacesList,
    serversList,
    isInterfacesLoading,
    isServersLoading,
  ])

  return {
    interfacesList,
    interfacesError,
    isInterfacesLoading,
    serversList,
    serversError,
    isServersLoading,
  }
}
