'use client'

import {
  Control,
  FieldValues,
  FieldPath,
  UseFormSetValue,
} from 'react-hook-form'
import { Interface } from '@/schema/interfaces.ts'
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
import { ServerName } from '@/components/server-select'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  label: string
  control: Control<T>
  setValue: UseFormSetValue<T>
  options: Array<Interface>
  isLoading?: boolean
  error?: unknown
}

export function InterfaceSelect<T extends FieldValues>({
  name,
  label,
  control,
  options,
  isLoading,
  error,
}: Props<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Select
              onValueChange={(selectedId) => {
                const selected = options.find(
                  (opt) => opt.id.toString() === selectedId
                )
                if (selected) {
                  field.onChange(selected.id) // sets form field interface_id = selected.id
                }
              }}
              value={field.value?.toString()}
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
                      <SelectItem key={item.id} value={item.id.toString()}>
                        {item.name} — <ServerName id={item.server_id} />
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
