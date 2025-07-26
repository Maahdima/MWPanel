'use client'

import { Control, FieldValues, Path } from 'react-hook-form'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface InterfaceSelectProps<T extends FieldValues> {
  name: Path<T>
  label: string
  control: Control<T>
  setValue: (name: Path<T>, value: any) => void
  options: Array<{ interface_id: string; name: string; listen_port: string }>
  isLoading?: boolean
  error?: unknown
}

export function InterfaceSelect<T extends FieldValues>({
  name,
  label,
  control,
  setValue,
  options,
  isLoading,
  error,
}: InterfaceSelectProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Select
              onValueChange={(selectedName) => {
                field.onChange(selectedName)
                const selected = options.find(
                  (opt) => opt.name === selectedName
                )
                if (selected) {
                  setValue('interface_id' as Path<T>, selected.interface_id)
                  setValue('listen_port' as Path<T>, selected.listen_port)
                }
              }}
              value={field.value}
            >
              <SelectTrigger className='w-full'>
                <SelectValue placeholder='Select interface' />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>Interfaces</SelectLabel>
                  {isLoading ? (
                    <SelectItem value='loading' disabled>
                      Loading...
                    </SelectItem>
                  ) : error ? (
                    <SelectItem value='error' disabled>
                      Error loading interfaces
                    </SelectItem>
                  ) : options.length === 0 ? (
                    <SelectItem value='empty' disabled>
                      No interfaces available
                    </SelectItem>
                  ) : (
                    options.map((item) => (
                      <SelectItem key={item.name} value={item.name}>
                        {item.name}
                      </SelectItem>
                    ))
                  )}
                </SelectGroup>
              </SelectContent>
            </Select>
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
