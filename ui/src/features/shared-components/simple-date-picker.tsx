import { useState } from 'react'
import { format } from 'date-fns'
import { CalendarIcon, XIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

type SimpleDatepickerProps = {
  value: string | null | undefined
  onChange: (value: string | null) => void
  placeholder?: string
  minDate?: Date
}

export function SimpleDatepicker({
  value,
  onChange,
  placeholder = 'Pick a date (optional)',
  minDate = new Date(),
}: SimpleDatepickerProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <div className='relative w-full'>
        <PopoverTrigger asChild>
          <Button
            variant='outline'
            className={cn(
              'w-full justify-start pr-10 pl-3 text-left font-normal',
              !value && 'text-muted-foreground'
            )}
          >
            {value ? format(new Date(value), 'yyyy-MM-dd') : placeholder}
            <CalendarIcon className='ml-auto h-4 w-4 opacity-50' />
          </Button>
        </PopoverTrigger>

        {value && (
          <Button
            type='button'
            onClick={() => onChange(null)}
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
          hidden={{ before: minDate }}
          selected={value ? new Date(value) : undefined}
          onSelect={(date) => {
            if (date) {
              onChange(format(new Date(date), 'yyyy-MM-dd'))
            }
            setIsOpen(false)
          }}
        />
      </PopoverContent>
    </Popover>
  )
}
