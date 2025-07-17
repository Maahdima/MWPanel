import ContentSection from '../components/content-section'
import { AccountForm } from './account-form'

export default function SettingsAccount() {
  return (
    <ContentSection title='Account' desc='Update your account credentials.'>
      <AccountForm />
    </ContentSection>
  )
}
