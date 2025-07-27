'use client'

import { Control, UseFormSetValue } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers.ts'
import { fetchPeerCredentials } from '@/api/peers.ts'
import { Button } from '@/components/ui/button.tsx'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form.tsx'
import { Input } from '@/components/ui/input.tsx'
import { PasswordInput } from '@/components/password-input.tsx'

interface Props {
  control: Control<CreatePeerRequest>
  setValue: UseFormSetValue<CreatePeerRequest>
}

export function GenerateKeyField({ control, setValue }: Props) {
  const onGenerateKeys = async () => {
    const { private_key, public_key, allowed_address } =
      await fetchPeerCredentials()
    setValue('private_key', private_key)
    setValue('public_key', public_key)
    setValue('allowed_address', allowed_address)
  }

  return (
    <>
      <div className='flex items-end gap-2'>
        <FormField
          control={control}
          name='private_key'
          render={({ field }) => (
            <FormItem className='w-full'>
              <FormLabel>Private Key</FormLabel>
              <FormControl>
                <PasswordInput
                  className='bg-input/30'
                  {...field}
                  placeholder='Private key'
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type='button' onClick={onGenerateKeys} variant='outline'>
          Generate
        </Button>
      </div>

      <FormField
        control={control}
        name='public_key'
        render={({ field }) => (
          <FormItem>
            <FormLabel>Public Key</FormLabel>
            <FormControl>
              <Input {...field} placeholder='Public key' />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </>
  )
}
