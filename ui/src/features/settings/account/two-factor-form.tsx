import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import {
  totpConfirmRequestSchema,
  totpDisableRequestSchema,
  TotpConfirmRequest,
  TotpDisableRequest,
  TotpSetup,
} from '@/schema/authentication.ts'
import { useTotpStatusQuery } from '@/hooks/authentication/useTotpStatusQuery.tsx'
import {
  useConfirmTotpSetupMutation,
  useDisableTotpMutation,
  useStartTotpSetupMutation,
} from '@/hooks/authentication/useTotpMutations.tsx'
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
import { PasswordInput } from '@/components/password-input'
import { Skeleton } from '@/components/ui/skeleton'

export function TwoFactorForm() {
  const { data: status, isLoading } = useTotpStatusQuery()
  const startSetup = useStartTotpSetupMutation()
  const confirmSetup = useConfirmTotpSetupMutation()
  const disableTotp = useDisableTotpMutation()

  const [setup, setSetup] = useState<TotpSetup | null>(null)
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null)
  const [showDisable, setShowDisable] = useState(false)

  const confirmForm = useForm<TotpConfirmRequest>({
    resolver: zodResolver(totpConfirmRequestSchema),
    defaultValues: { code: '' },
  })

  const disableForm = useForm<TotpDisableRequest>({
    resolver: zodResolver(totpDisableRequestSchema),
    defaultValues: { password: '', code: '' },
  })

  async function handleStartSetup() {
    const result = await startSetup.mutateAsync()
    setSetup(result)
    setRecoveryCodes(null)
    confirmForm.reset()
  }

  async function handleConfirm(data: TotpConfirmRequest) {
    const result = await confirmSetup.mutateAsync(data)
    setSetup(null)
    setRecoveryCodes(result.recovery_codes)
  }

  async function handleDisable(data: TotpDisableRequest) {
    await disableTotp.mutateAsync(data)
    setShowDisable(false)
    disableForm.reset()
  }

  function copyRecoveryCodes() {
    if (!recoveryCodes) return
    navigator.clipboard.writeText(recoveryCodes.join('\n'))
    toast.success('Recovery codes copied.')
  }

  if (isLoading) {
    return <Skeleton className='h-24 w-full' />
  }

  if (recoveryCodes) {
    return (
      <div className='space-y-4'>
        <div>
          <h4 className='font-medium'>Save your recovery codes</h4>
          <p className='text-muted-foreground text-sm'>
            Store these somewhere safe. Each code can be used once if you lose
            access to your authenticator app.
          </p>
        </div>
        <ul className='bg-muted grid gap-1 rounded-md p-3 font-mono text-sm'>
          {recoveryCodes.map((code) => (
            <li key={code}>{code}</li>
          ))}
        </ul>
        <div className='flex gap-2'>
          <Button type='button' onClick={copyRecoveryCodes}>
            Copy codes
          </Button>
          <Button type='button' variant='outline' onClick={() => setRecoveryCodes(null)}>
            Done
          </Button>
        </div>
      </div>
    )
  }

  if (setup) {
    return (
      <div className='space-y-4'>
        <div>
          <h4 className='font-medium'>Scan QR code</h4>
          <p className='text-muted-foreground text-sm'>
            Use Google Authenticator, Authy, or any TOTP app, then enter the
            6-digit code to confirm.
          </p>
        </div>
        <img
          src={setup.qr_code_data_url}
          alt='TOTP QR code'
          className='border-border size-48 rounded-md border bg-white p-2'
        />
        <p className='text-muted-foreground text-xs break-all'>
          Manual secret: <span className='font-mono'>{setup.secret}</span>
        </p>
        <Form {...confirmForm}>
          <form
            onSubmit={confirmForm.handleSubmit(handleConfirm)}
            className='space-y-3'
          >
            <FormField
              control={confirmForm.control}
              name='code'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Verification code</FormLabel>
                  <FormControl>
                    <Input placeholder='123456' inputMode='numeric' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className='flex gap-2'>
              <Button type='submit' disabled={confirmSetup.isPending}>
                Enable 2FA
              </Button>
              <Button
                type='button'
                variant='outline'
                onClick={() => setSetup(null)}
              >
                Cancel
              </Button>
            </div>
          </form>
        </Form>
      </div>
    )
  }

  if (status?.enabled) {
    return (
      <div className='space-y-4'>
        <p className='text-sm'>
          Two-factor authentication is <strong>enabled</strong>.
        </p>
        {showDisable ? (
          <Form {...disableForm}>
            <form
              onSubmit={disableForm.handleSubmit(handleDisable)}
              className='space-y-3'
            >
              <FormField
                control={disableForm.control}
                name='password'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <PasswordInput placeholder='Your password' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={disableForm.control}
                name='code'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Authenticator or recovery code</FormLabel>
                    <FormControl>
                      <Input placeholder='123456 or XXXX-XXXX' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className='flex gap-2'>
                <Button
                  type='submit'
                  variant='destructive'
                  disabled={disableTotp.isPending}
                >
                  Disable 2FA
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  onClick={() => setShowDisable(false)}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </Form>
        ) : (
          <Button
            type='button'
            variant='destructive'
            onClick={() => setShowDisable(true)}
          >
            Disable two-factor authentication
          </Button>
        )}
      </div>
    )
  }

  return (
    <div className='space-y-3'>
      <p className='text-muted-foreground text-sm'>
        Protect your admin account with an authenticator app (TOTP) and recovery
        codes.
      </p>
      <Button
        type='button'
        onClick={handleStartSetup}
        disabled={startSetup.isPending}
      >
        Set up two-factor authentication
      </Button>
    </div>
  )
}
