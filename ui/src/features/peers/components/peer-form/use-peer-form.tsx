'use client'

import { useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreatePeerRequest,
  CreatePeerSchema,
  UpdatePeerRequest,
  UpdatePeerSchema,
} from '@/schema/peers'
import { toast } from 'sonner'
import { fetchPeerCredentials } from '@/api/peers.ts'
import { useCreatePeerMutation } from '@/hooks/peers/useCreatePeerMutation'
import { useUpdatePeerMutation } from '@/hooks/peers/useUpdatePeerMutation'
import { usePeerDefaults } from '@/features/peers/components/peer-form/use-peer-defaults.tsx'

export function usePeerForm({
  currentRow,
  onClose,
}: {
  currentRow?: Partial<CreatePeerRequest>
  onClose: () => void
}) {
  const isEdit = !!currentRow

  const form = useForm<CreatePeerRequest>({
    resolver: zodResolver(
      (isEdit ? UpdatePeerSchema : CreatePeerSchema) as never
    ),
    defaultValues: (isEdit ? currentRow : {}) as never,
  })

  const [shouldRefetch, setShouldRefetch] = useState(false)
  const [isDefaultsReady, setIsDefaultsReady] = useState(isEdit)

  useEffect(() => {
    if (!isEdit) {
      setShouldRefetch(true)
    }
  }, [isEdit])

  const {
    interfacesList,
    isInterfacesLoading,
    interfacesError,
    serversList,
    isServersLoading,
    serversError,
  } = usePeerDefaults({ isEdit, form, shouldRefetch })

  const { mutateAsync: createPeer } = useCreatePeerMutation()
  const { mutateAsync: updatePeer } = useUpdatePeerMutation()

  const hasGeneratedKeys = useRef(false)

  useEffect(() => {
    const generateKeys = async () => {
      if (!isEdit && !hasGeneratedKeys.current) {
        hasGeneratedKeys.current = true
        const { private_key, public_key, allowed_address } =
          await fetchPeerCredentials()
        form.setValue('private_key', private_key)
        form.setValue('public_key', public_key)
        form.setValue('allowed_address', allowed_address)
        setIsDefaultsReady(true)
      }
    }

    generateKeys()
  }, [isEdit, form])

  useEffect(() => {
    if (isEdit && currentRow) {
      form.reset(currentRow)
      setIsDefaultsReady(true)
    }
  }, [isEdit, currentRow, form])

  const onSubmit = async (values: CreatePeerRequest | UpdatePeerRequest) => {
    if (isEdit) {
      await updatePeer(values as UpdatePeerRequest)
      toast.success('Peer updated successfully.', { duration: 5000 })
    } else {
      await createPeer(values as CreatePeerRequest)
      toast.success('Peer created successfully.', { duration: 5000 })
    }
    form.reset()
    onClose()
  }

  return {
    form,
    onSubmit,
    isDefaultsReady,
    interfacesList,
    isInterfacesLoading,
    interfacesError,
    serversList,
    isServersLoading,
    serversError,
  }
}
