import { HTMLAttributes } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from '@tanstack/react-router'
import { loginRequestSchema } from '@/schema/authentication.ts'
import { useAuthStore } from '@/stores/authStore.ts'
import { cn } from '@/lib/utils'
import { useLoginMutation } from '@/hooks/authentication/useLoginMutation.tsx'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox.tsx'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label.tsx'
import { PasswordInput } from '@/components/password-input'

type UserAuthFormProps = HTMLAttributes<HTMLFormElement>

export function UserAuthForm({ className, ...props }: UserAuthFormProps) {
  const authStore = useAuthStore()
  const router = useRouter()

  const { mutateAsync: loginMutation, isPending: isLoginPending } =
    useLoginMutation()

  const form = useForm<z.infer<typeof loginRequestSchema>>({
    resolver: zodResolver(loginRequestSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  async function onSubmit(data: z.infer<typeof loginRequestSchema>) {
    const response = await loginMutation(data)

    if (response) {
      const admin = {
        user_id: response.user_id,
        username: response.username,
      }

      authStore.auth.setAccessToken(response.access_token)
      authStore.auth.setAdmin(admin)
      
      form.reset()

      router.navigate({ to: '/' })
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder='username' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem className='relative'>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className='flex items-center gap-3'>
          <Checkbox id='remember' />
          <Label htmlFor='remember'>Remember me</Label>
        </div>
        <Button className='mt-2' disabled={isLoginPending}>
          Login
        </Button>
      </form>
    </Form>
  )
}
