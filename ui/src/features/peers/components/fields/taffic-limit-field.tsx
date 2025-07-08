'use client'

import { useEffect, useState } from 'react'
import { Control } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers.ts'
import { XIcon } from 'lucide-react'
import { Button } from '@/components/ui/button.tsx'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form.tsx'
import { Input } from '@/components/ui/input.tsx'

interface Props {
  control: Control<CreatePeerRequest>
}

export function TrafficInput({ control }: Props) {
  const [rawValue, setRawValue] = useState('')

  return (
    <FormField
      control={control}
      name='traffic_limit'
      render={({ field }) => {
        useEffect(() => {
          if (field.value !== null && field.value !== undefined) {
            setRawValue(field.value.toString())
          } else {
            setRawValue('')
          }
        }, [field.value])

        return (
          <FormItem>
            <FormLabel>Traffic Limit</FormLabel>
            <FormControl>
              <div className='relative'>
                <Input
                  type='number'
                  min={0}
                  id='traffic_limit'
                  value={rawValue}
                  placeholder='e.g., 1 (optional)'
                  onChange={(e) => {
                    const val = e.target.value
                    setRawValue(val)
                    field.onChange(val ? val : null)
                  }}
                />

                {rawValue && (
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon'
                    className='absolute top-1/2 right-10 h-7 w-7 -translate-y-1/2 text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100'
                    onClick={() => {
                      setRawValue('')
                      field.onChange(null)
                    }}
                  >
                    <XIcon className='h-4 w-4' />
                    <span className='sr-only'>Clear</span>
                  </Button>
                )}

                <div className='text-muted-foreground bg-muted border-input absolute inset-y-0 right-0 flex items-center rounded-none rounded-r-md border-l px-2 text-sm'>
                  GB
                </div>
              </div>
            </FormControl>
            <FormMessage />
          </FormItem>
        )
      }}
    />
  )
}
