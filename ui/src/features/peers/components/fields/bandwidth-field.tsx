'use client'

import { useEffect, useState } from 'react'
import { Control, useController } from 'react-hook-form'
import { CreatePeerRequest } from '@/schema/peers'
import { XIcon } from 'lucide-react'
import { Button } from '@/components/ui/button.tsx'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface Props {
  control: Control<CreatePeerRequest>
  name: 'upload_bandwidth' | 'download_bandwidth'
  label: string
}

const units = ['Kb', 'Mb', 'Gb']

export function BandwidthField({ control, name, label }: Props) {
  const { field: watchedField } = useController({ control, name })
  const [rawValue, setRawValue] = useState('')
  const [unit, setUnit] = useState('Mb')

  const getCombinedValue = (val: string, unit: string) => {
    if (!val) return ''
    const suffix = unit === 'Kb' ? 'K' : unit === 'Mb' ? 'M' : 'G'
    return `${val}${suffix}`
  }

  useEffect(() => {
    if (!watchedField.value) return
    const match = watchedField.value.match(/^(\d+(?:\.\d+)?)([KMG])$/i)
    if (match) {
      setRawValue(match[1])
      const unitMap = { K: 'Kb', M: 'Mb', G: 'Gb' }
      setUnit(unitMap[match[2].toUpperCase() as 'K' | 'M' | 'G'] ?? 'Mb')
    }
  }, [watchedField.value])

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        return (
          <FormItem>
            <FormLabel>{label}</FormLabel>
            <div className='flex min-w-0 gap-2'>
              <FormControl>
                <div className='relative w-full'>
                  <Input
                    placeholder='e.g. 24 (optional)'
                    value={rawValue}
                    onChange={(e) => {
                      const val = e.target.value
                      setRawValue(val)
                      field.onChange(val ? getCombinedValue(val, unit) : null)
                    }}
                  />
                  {field.value && (
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      className='absolute top-1/2 right-1 h-7 w-7 -translate-y-1/2 text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100'
                      onClick={() => {
                        setRawValue('')
                        field.onChange(null)
                      }}
                    >
                      <XIcon className='h-4 w-4' />
                      <span className='sr-only'>Clear</span>
                    </Button>
                  )}
                </div>
              </FormControl>
              <Select
                value={unit}
                onValueChange={(newUnit) => {
                  setUnit(newUnit)
                  field.onChange(getCombinedValue(rawValue, newUnit))
                }}
              >
                <SelectTrigger className='w-[72px] shrink-0'>
                  <SelectValue placeholder='Unit' />
                </SelectTrigger>
                <SelectContent>
                  {units.map((u) => (
                    <SelectItem key={u} value={u}>
                      {u}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <FormMessage />
          </FormItem>
        )
      }}
    />
  )
}
