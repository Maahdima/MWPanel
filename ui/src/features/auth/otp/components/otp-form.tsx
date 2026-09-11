import { HTMLAttributes, useEffect, useRef, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { TEMP_2FA_TOKEN_KEY } from '@/lib/auth-2fa'
import { useAuthStore } from '@/stores/authStore.ts'
import { cn } from '@/lib/utils'
import { useVerify2FAMutation } from '@/hooks/authentication/useVerify2FAMutation.tsx'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
  InputOTPSeparator,
} from '@/components/ui/input-otp'

type OtpFormProps = HTMLAttributes<HTMLFormElement>

const totpSchema = z.object({
  otp: z.string().length(6, { message: 'Enter the 6-digit code from your app.' }),
})

const recoverySchema = z.object({
  code: z.string().min(8, { message: 'Enter a valid recovery code.' }),
})

export function OtpForm({ className, ...props }: OtpFormProps) {
  const navigate = useNavigate()
  const authStore = useAuthStore()
  const [useRecovery, setUseRecovery] = useState(false)
  const submittingRef = useRef(false)
  const { mutateAsync: verify2FA, isPending } = useVerify2FAMutation()

  useEffect(() => {
    const tempToken = sessionStorage.getItem(TEMP_2FA_TOKEN_KEY)
    if (!tempToken) {
      navigate({ to: '/sign-in' })
    }
  }, [navigate])

  const totpForm = useForm<z.infer<typeof totpSchema>>({
    resolver: zodResolver(totpSchema),
    defaultValues: { otp: '' },
  })

  const recoveryForm = useForm<z.infer<typeof recoverySchema>>({
    resolver: zodResolver(recoverySchema),
    defaultValues: { code: '' },
  })

  async function completeLogin(code: string) {
    if (submittingRef.current) return
    submittingRef.current = true

    const tempToken = sessionStorage.getItem(TEMP_2FA_TOKEN_KEY)
    if (!tempToken) {
      submittingRef.current = false
      navigate({ to: '/sign-in' })
      return
    }

    try {
      const response = await verify2FA({ temp_token: tempToken, code })
      if (!response.access_token || !response.user_id || !response.username) {
        submittingRef.current = false
        return
      }

      sessionStorage.removeItem(TEMP_2FA_TOKEN_KEY)
      authStore.auth.setAccessToken(response.access_token)
      authStore.auth.setAdmin({
        user_id: response.user_id,
        username: response.username,
      })
      navigate({ to: '/' })
    } catch {
      submittingRef.current = false
      totpForm.setValue('otp', '')
    }
  }

  async function onRecoverySubmit(data: z.infer<typeof recoverySchema>) {
    await completeLogin(data.code)
  }

  if (useRecovery) {
    return (
      <Form {...recoveryForm}>
        <form
          onSubmit={recoveryForm.handleSubmit(onRecoverySubmit)}
          className={cn('grid gap-3', className)}
          {...props}
        >
          <FormField
            control={recoveryForm.control}
            name='code'
            render={({ field }) => (
              <FormItem>
                <FormLabel>Recovery code</FormLabel>
                <FormControl>
                  <Input placeholder='XXXX-XXXX' autoComplete='off' {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button className='mt-2' disabled={isPending}>
            Verify
          </Button>
          <Button
            type='button'
            variant='ghost'
            onClick={() => setUseRecovery(false)}
          >
            Use authenticator code
          </Button>
        </form>
      </Form>
    )
  }

  return (
    <Form {...totpForm}>
      <form
        onSubmit={(e) => e.preventDefault()}
        className={cn('grid gap-2', className)}
        {...props}
      >
        <FormField
          control={totpForm.control}
          name='otp'
          render={({ field }) => (
            <FormItem>
              <FormLabel className='sr-only'>One-Time Password</FormLabel>
              <FormControl>
                <InputOTP
                  maxLength={6}
                  value={field.value}
                  disabled={isPending}
                  onChange={(value) => {
                    field.onChange(value)
                    if (value.length === 6) {
                      void completeLogin(value)
                    }
                  }}
                  containerClassName='justify-between sm:[&>[data-slot="input-otp-group"]>div]:w-12'
                >
                  <InputOTPGroup>
                    <InputOTPSlot index={0} />
                    <InputOTPSlot index={1} />
                    <InputOTPSlot index={2} />
                  </InputOTPGroup>
                  <InputOTPSeparator />
                  <InputOTPGroup>
                    <InputOTPSlot index={3} />
                    <InputOTPSlot index={4} />
                    <InputOTPSlot index={5} />
                  </InputOTPGroup>
                </InputOTP>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button
          type='button'
          variant='ghost'
          disabled={isPending}
          onClick={() => setUseRecovery(true)}
        >
          Use a recovery code
        </Button>
      </form>
    </Form>
  )
}
