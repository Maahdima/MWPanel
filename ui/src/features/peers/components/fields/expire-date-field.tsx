'use client'

import { Control } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers'
import {
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { SimpleDatepicker } from '@/features/shared-components/simple-date-picker.tsx'

interface Props {
  control: Control<CreatePeerRequest>
}

export function ExpireDateField({ control }: Props) {
  return (
    <FormField
      control={control}
      name='expire_time'
      render={({ field }) => (
        <FormItem className='flex flex-col'>
          <FormLabel>Expire Time</FormLabel>
          <SimpleDatepicker value={field.value} onChange={field.onChange} />
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
