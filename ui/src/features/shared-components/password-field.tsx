import { Control, FieldValues, FieldPath } from 'react-hook-form'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form.tsx'
import { PasswordInput } from '@/components/password-input.tsx'

interface Props<T extends FieldValues> {
  name: FieldPath<T>
  label: string
  placeholder?: string
  control: Control<T>
}

export function PasswordField<T extends FieldValues>({
  name,
  label,
  placeholder,
  control,
}: Props<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <PasswordInput
              className='bg-input/30'
              {...field}
              placeholder={placeholder}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  )
}
