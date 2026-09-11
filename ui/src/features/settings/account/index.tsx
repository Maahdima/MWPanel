import ContentSection from '../components/content-section'
import { AccountForm } from './account-form'
import { TwoFactorForm } from './two-factor-form'
import { Separator } from '@/components/ui/separator'

export default function SettingsAccount() {
  return (
    <ContentSection title='Account' desc='Update your account credentials and security.'>
      <div className='space-y-8'>
        <AccountForm />
        <div>
          <h4 className='mb-1 text-base font-medium'>Two-factor authentication</h4>
          <p className='text-muted-foreground mb-4 text-sm'>
            Add an extra layer of security to your admin login.
          </p>
          <Separator className='mb-4' />
          <TwoFactorForm />
        </div>
      </div>
    </ContentSection>
  )
}
