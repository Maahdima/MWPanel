import { Control } from 'react-hook-form';
import { Dices } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form.tsx';
import { Input } from '@/components/ui/input.tsx'


interface Props {
  name: string
  label: string
  placeholder?: string
  control: Control<any>
}

export function SimpleField({ name, label, placeholder, control }: Props) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        const { value, onChange, ...rest } = field
        return (
          <FormItem>
            <FormLabel>{label}</FormLabel>
            <FormControl>
              {name === 'name' ? (
                <div className='flex items-center gap-2'>
                  <Input
                    className='flex-1'
                    placeholder={placeholder}
                    {...rest}
                    value={String(value ?? '')}
                    onChange={(e) => onChange(e.target.value)}
                  />
                  <Button
                    type='button'
                    variant='outline'
                    size='icon'
                    onClick={() => {
                      const chars =
                        'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
                      let s = ''
                      for (let i = 0; i < 10; i++) {
                        s += chars.charAt(Math.floor(Math.random() * chars.length))
                      }
                      onChange(s)
                      // focus the input after generating
                      setTimeout(() => {
                        const el = document.querySelector<HTMLInputElement>(
                          `input[name="${name}"]`
                        )
                        el?.focus()
                        el?.select()
                      }, 0)
                    }}
                    aria-label='Generate random name'
                  >
                    <Dices  />
                  </Button>
                </div>
              ) : (
                <Input
                  placeholder={placeholder}
                  {...rest}
                  value={String(value ?? '')}
                  onChange={(e) => onChange(e.target.value)}
                />
              )}
            </FormControl>
            <FormMessage />
          </FormItem>
        )
      }}
    />
  )
}