import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  UpdateProfileRequest,
  updateProfileSchema,
} from '@/schema/authentication.ts'
import { useUpdateProfileMutation } from '@/hooks/authentication/useUpdateProfileMutation.tsx'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordField } from '@/features/shared-components/password-field.tsx'

export function AccountForm() {
  const updateProfile = useUpdateProfileMutation()

  const form = useForm<UpdateProfileRequest>({
    resolver: zodResolver(updateProfileSchema),
  })

  function onSubmit(data: UpdateProfileRequest) {
    updateProfile.mutate(data)
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-4'>
        <FormField
          control={form.control}
          name='old_username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Old Username</FormLabel>
              <FormControl>
                <Input placeholder='Your username (Required)' {...field} />
              </FormControl>
              <FormDescription>
                This is the name that will be used to login into the panel.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <PasswordField
          name='old_password'
          label='Old Password'
          control={form.control}
          placeholder='Your password (Required)'
        />
        <FormField
          control={form.control}
          name='new_username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>New Username</FormLabel>
              <FormControl>
                <Input placeholder='Your new username (optional)' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <PasswordField
          name='new_password'
          label='New Password'
          control={form.control}
          placeholder='Your new password (optional)'
        />
        <Button type='submit'>Update account</Button>
      </form>
    </Form>
  )
}
