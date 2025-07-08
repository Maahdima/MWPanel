'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { Control } from 'react-hook-form'
import { CalendarIcon } from '@radix-ui/react-icons'
import { CreatePeerRequest } from '@/schema/peers'
import { XIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

interface Props {
  control: Control<CreatePeerRequest>
}

export function ExpireDateField({ control }: Props) {
  const [isCalendarOpen, setIsCalendarOpen] = useState(false)

  return (
    <FormField
      control={control}
      name='expire_time'
      render={({ field }) => (
        <FormItem className='flex flex-col'>
          <FormLabel>Expire Time</FormLabel>
          <Popover open={isCalendarOpen} onOpenChange={setIsCalendarOpen}>
            <div className='relative w-full'>
              <PopoverTrigger asChild>
                <FormControl>
                  <Button
                    variant='outline'
                    className={cn(
                      'w-full justify-start pr-10 pl-3 text-left font-normal',
                      !field.value && 'text-muted-foreground'
                    )}
                  >
                    {field.value
                      ? format(new Date(field.value), 'yyyy-MM-dd')
                      : 'Pick a date (optional)'}
                    <CalendarIcon className='ml-auto h-4 w-4 opacity-50' />
                  </Button>
                </FormControl>
              </PopoverTrigger>
              {field.value && (
                <Button
                  type='button'
                  onClick={() => field.onChange(null)}
                  className='hover:bg-accent/50 absolute top-1/2 right-8 h-7 w-7 -translate-y-1/2 bg-transparent text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100'
                >
                  <XIcon className='h-4 w-4' />
                  <span className='sr-only'>Clear date</span>
                </Button>
              )}
            </div>

            <PopoverContent className='rounded-md p-0 shadow-md'>
              <Calendar
                mode='single'
                hidden={{ before: new Date() }}
                selected={field.value ? new Date(field.value) : undefined}
                onSelect={(date) => {
                  if (date) {
                    field.onChange(format(new Date(date), 'yyyy-MM-dd'))
                  }
                  setIsCalendarOpen(false)
                }}
              />
            </PopoverContent>
          </Popover>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
