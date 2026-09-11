import ContentSection from '../components/content-section'
import { TwoFactorForm } from './two-factor-form'

export default function SettingsSecurity() {
  return (
    <ContentSection
      title='Security'
      desc='Protect your admin account with two-factor authentication.'
    >
      <TwoFactorForm />
    </ContentSection>
  )
}
