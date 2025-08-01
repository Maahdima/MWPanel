'use client'

import { useEffect, useRef } from 'react'
import { UseFormReturn, useWatch } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { usePeerAllowedAddressMutation } from '@/hooks/peers/usePeerAllowedAddressMutation.ts'
import { useServersListQuery } from '@/hooks/servers/useServersListQuery'

interface Props {
  isEdit: boolean
  form: UseFormReturn<CreatePeerRequest>
  shouldRefetch: boolean
}

export function usePeerDefaults({ isEdit, form, shouldRefetch }: Props) {
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

  const { mutate: GetPeerAllowedAddress } = usePeerAllowedAddressMutation()

  const hasSetDefaults = useRef(false)

  // Watch for interface_id changes
  const interfaceId = useWatch({
    control: form.control,
    name: 'interface_id',
  })

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

  useEffect(() => {
    if (interfaceId && !isEdit) {
      GetPeerAllowedAddress(
        { interface_id: interfaceId },
        {
          onSuccess: (data) => {
            form.setValue('allowed_address', data.allowed_address)
          },
        }
      )
    }
  }, [interfaceId, GetPeerAllowedAddress, form, isEdit])

  return {
    interfacesList,
    interfacesError,
    isInterfacesLoading,
    serversList,
    serversError,
    isServersLoading,
  }
}
