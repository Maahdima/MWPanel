'use client'

import { Control, UseFormSetValue } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery'
import { BandwidthField } from '@/features/peers/components/fields/bandwidth-field'
import { ExpireDateField } from '@/features/peers/components/fields/expire-date-field'
import { GenerateKeyField } from '@/features/peers/components/fields/generate-key-field'
import { InterfaceSelect } from '@/features/peers/components/fields/interface-select'
import { TrafficInput } from '@/features/peers/components/fields/taffic-limit-field'
import { SimpleField } from '@/features/shared-components/simple-field'

interface PeerFormFieldsProps {
  form: {
    control: Control<CreatePeerRequest>
    setValue: UseFormSetValue<CreatePeerRequest>
  }
  isEdit: boolean
}

export function PeerFormFields({ form, isEdit }: PeerFormFieldsProps) {
  const {
    data: interfacesList = [],
    isLoading: isInterfacesLoading,
    error: interfacesError,
  } = useInterfacesListQuery()

  return (
    <>
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
        <>
          <InterfaceSelect
            name='interface_name'
            label='Interface'
            control={form.control}
            setValue={form.setValue}
            options={interfacesList}
            isLoading={isInterfacesLoading}
            error={interfacesError}
          />
          <GenerateKeyField control={form.control} setValue={form.setValue} />
          <SimpleField
            name='endpoint'
            label='Endpoint'
            placeholder='e.g., 185.51.200.10'
            control={form.control}
          />
        </>
      )}

      <SimpleField
        name='allowed_address'
        label='Allowed Address'
        placeholder='e.g., 10.0.0.2/32'
        control={form.control}
      />
      <SimpleField
        name='persistent_keepalive'
        label='Persistent Keepalive'
        placeholder='e.g., 00:00:25 (optional)'
        control={form.control}
      />
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
    </>
  )
}
