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
  control: Control<CreatePeerRequest>
  setValue: UseFormSetValue<CreatePeerRequest>
  isEdit: boolean
}

export function PeerFormFields({
  control,
  setValue,
  isEdit,
}: PeerFormFieldsProps) {
  const {
    data: interfacesList = [],
    isLoading: isInterfacesLoading,
    error: interfacesError,
  } = useInterfacesListQuery()

  return (
    <div className='grid grid-cols-1 gap-x-3 gap-y-4 md:grid-cols-2'>
      <SimpleField
        name='name'
        label='Name'
        placeholder='Peer Name'
        control={control}
      />
      <SimpleField
        name='comment'
        label='Comment'
        placeholder='Comment (optional)'
        control={control}
      />

      {!isEdit && (
        <div className='md:col-span-2'>
          <div className='grid grid-cols-1 gap-x-3 gap-y-4 md:grid-cols-2'>
            <InterfaceSelect
              name='interface_id'
              label='Interface'
              control={control}
              setValue={setValue}
              options={interfacesList}
              isLoading={isInterfacesLoading}
              error={interfacesError}
            />
            <SimpleField
              name='endpoint'
              label='Endpoint'
              placeholder='e.g., 185.51.200.10'
              control={control}
            />
            <div className='md:col-span-2'>
              <GenerateKeyField control={control} setValue={setValue} />
            </div>
          </div>
        </div>
      )}

      <SimpleField
        name='allowed_address'
        label='Allowed Address'
        placeholder='e.g., 10.0.0.2/32'
        control={control}
      />
      <SimpleField
        name='persistent_keepalive'
        label='Persistent Keepalive'
        placeholder='e.g., 00:00:25 (optional)'
        control={control}
      />

      <TrafficInput control={control} />
      <ExpireDateField control={control} />

      <div className='space-y-4 md:col-span-2'>
        <BandwidthField
          name='download_bandwidth'
          label='Download Bandwidth'
          control={control}
        />
        <BandwidthField
          name='upload_bandwidth'
          label='Upload Bandwidth'
          control={control}
        />
      </div>
    </div>
  )
}
