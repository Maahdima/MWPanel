'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreatePeerRequest,
  CreatePeerSchema,
  UpdatePeerRequest,
  UpdatePeerSchema,
} from '@/schema/peers'
import { toast } from 'sonner'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { useCreatePeerMutation } from '@/hooks/peers/useCreatePeerMutation'
import { useUpdatePeerMutation } from '@/hooks/peers/useUpdatePeerMutation.ts'
import { Form } from '@/components/ui/form'
import { BandwidthField } from '@/features/peers/components/fields/bandwidth-field'
import { ExpireDateField } from '@/features/peers/components/fields/expire-date-field'
import { GenerateKeyField } from '@/features/peers/components/fields/generate-key-field'
import { InterfaceSelect } from '@/features/peers/components/fields/interface-select'
import { TrafficInput } from '@/features/peers/components/fields/taffic-limit-field'
import { SimpleField } from '@/features/shared-components/simple-field.tsx'

interface PeerFormProps {
  currentRow?: Partial<CreatePeerRequest>
  onClose: () => void
}

export function PeerForm({ currentRow, onClose }: PeerFormProps) {
  const isEdit = Boolean(currentRow)
  const { mutateAsync: createPeer } = useCreatePeerMutation()
  const { mutateAsync: updatePeer } = useUpdatePeerMutation()

  const {
    data: interfacesList,
    isLoading: isInterfacesListLoading,
    error: interfacesListError,
  } = useInterfacesListQuery()

  const form = useForm<CreatePeerRequest>({
    resolver: zodResolver(
      (isEdit ? UpdatePeerSchema : CreatePeerSchema) as never
    ),
    defaultValues: currentRow ?? {},
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    }
  }, [currentRow, form])

  const onSubmit = async (values: CreatePeerRequest | UpdatePeerRequest) => {
    if (isEdit) {
      await updatePeer(values as UpdatePeerRequest)
    } else {
      await createPeer(values as CreatePeerRequest)
    }
    form.reset()
    const toastMessage = isEdit
      ? 'Peer updated successfully.'
      : 'Peer created successfully.'
    toast.success(toastMessage, { duration: 5000 })
    onClose()
  }

  return (
    <Form {...form}>
      <form
        id='peer-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className='space-y-4'
      >
        <SimpleField
          name='comment'
          label='Comment'
          placeholder='Comment (optional)'
          control={form.control}
        />
        <SimpleField
          name='name'
          label='Name'
          placeholder='Peer Name'
          control={form.control}
        />
        {!isEdit && (
          <InterfaceSelect
            name='interface_name'
            label='Interface'
            control={form.control}
            setValue={form.setValue}
            options={interfacesList ?? []}
            isLoading={isInterfacesListLoading}
            error={interfacesListError}
          />
        )}
        {!isEdit && (
          <GenerateKeyField control={form.control} setValue={form.setValue} />
        )}
        {[
          {
            name: 'allowed_address',
            label: 'Allowed Address',
            placeholder: 'e.g., 10.0.0.2/32',
          },
          {
            name: 'persistent_keepalive',
            label: 'Persistent Keepalive',
            placeholder: 'e.g., 00:00:25 (optional)',
          },
        ].map(({ name, label, placeholder }) => (
          <SimpleField
            key={name}
            name={name as keyof CreatePeerRequest}
            label={label}
            placeholder={placeholder}
            control={form.control}
          />
        ))}
        {!isEdit && (
          <SimpleField
            name='endpoint'
            label='Endpoint'
            placeholder='e.g., 185.51.200.10'
            control={form.control}
          />
        )}
        <TrafficInput control={form.control} />
        <ExpireDateField control={form.control} />
        <BandwidthField
          name='download_bandwidth'
          label='Download Bandwidth'
          control={form.control}
        />
        <BandwidthField
          name='upload_bandwidth'
          label='Upload Bandwidth'
          control={form.control}
        />
      </form>
    </Form>
  )
}
