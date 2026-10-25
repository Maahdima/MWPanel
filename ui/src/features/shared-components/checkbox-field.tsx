import { Control, FieldValues, FieldPath } from 'react-hook-form'
import { Checkbox } from '@/components/ui/checkbox'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  label: string
  control: Control<T>
}

export function CheckboxField<T extends FieldValues>({
  name,
  label,
  control,
}: Props<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className='flex flex-row items-start space-y-0 space-x-3'>
          <FormControl>
            <Checkbox
              checked={field.value}
              onCheckedChange={field.onChange}
              onBlur={field.onBlur}
              ref={field.ref}
            />
          </FormControl>
          <div className='space-y-1 leading-none'>
            <FormLabel>{label}</FormLabel>
            <FormMessage />
          </div>
        </FormItem>
      )}
    />
  )
}
