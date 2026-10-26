import { BillingPanel } from '@/features/resellers/components/billing-panel'
import { PageFrame } from '@/features/resellers/page-frame'

export default function BillingPage() {
  return (
    <PageFrame
      title='Billing'
      description='Balance, invoices and payment receipts'
    >
      <BillingPanel />
    </PageFrame>
  )
}
