import { z } from 'zod'

export const integerText = z.string().regex(/^-?\d+$/)

export const PrincipalSchema = z.object({
  account_id: z.number(),
  username: z.string(),
  role: z.enum(['superadmin', 'reseller']),
  reseller_id: z.number().optional(),
})

export const ResellerSchema = z.object({
  disabled_peers: z.number().default(0),
  expired_peers: z.number().default(0),
  provisioning_peers: z.number().default(0),
  failed_operations: z.number().default(0),
  sync_pending_peers: z.number().default(0),
  sync_error_peers: z.number().default(0),
  migration_id: z.string().nullable(),
  active_peers: z.number(),
  total_peers: z.number(),
  id: z.number(),
  name: z.string(),
  username: z.string(),
  is_active: z.boolean(),
  billing_mode: z.enum(['prepaid', 'postpaid']),
  admin_suspended: z.boolean(),
  billing_suspended: z.boolean(),
  next_invoice_at: z.number(),
  interval_days: z.number(),
  grace_days: z.number(),
  recurring_fee_rial: integerText,
  balance_rial: integerText,
  credit_rial: integerText,
  debt_rial: integerText,
  reserved_rial: integerText,
  rate_rial: integerText.nullable(),
  default_interface_id: z.number().nullable(),
  max_active_peers: integerText.nullable(),
  max_peer_traffic_bytes: integerText.nullable(),
  max_total_traffic_bytes: integerText.nullable(),
  max_duration_days: integerText.nullable(),
  resources: z.array(
    z.object({
      interface_id: z.number(),
      name: z.string(),
      server_id: z.number(),
      pool_id: z.number().nullable(),
    })
  ),
})

export const InstructionsSchema = z.object({
  rate_rial: integerText,
  card_number: z.string(),
  cardholder: z.string(),
  instructions: z.string(),
})

export const ReceiptSchema = z.object({
  id: z.number(),
  reseller_id: z.number(),
  amount_rial: integerText,
  paid_at: z.number(),
  transaction_reference: z.string(),
  state: z.enum(['pending', 'approved', 'rejected']),
  review_reason: z.string(),
  reviewed_at: z.number().nullable(),
  created_at: z.number(),
})

export const BillingBookSchema = z.object({
  reseller_id: z.number(),
  billing_mode: z.enum(['prepaid', 'postpaid']),
  rate_rial: integerText,
  reserved_rial: integerText,
  interval_days: z.number(),
  grace_days: z.number(),
  recurring_fee_rial: integerText,
  state: z.object({
    status: z.string(),
    balance_rial: integerText,
    available_rial: integerText,
    overdue_rial: integerText,
    billable_blocked: z.boolean(),
    billing_suspended: z.boolean(),
    administrative_suspended: z.boolean(),
    due_at: z.number(),
    grace_until: z.number(),
    next_invoice_at: z.number(),
    activation_anchor: z.number(),
    reconciling: z.boolean(),
  }),
  recent_invoices: z.array(
    z.object({
      id: z.number(),
      period_start: z.number(),
      period_end: z.number(),
      due_at: z.number(),
      grace_until: z.number(),
      total_rial: integerText,
      due_rial: integerText,
      remaining_rial: integerText,
    })
  ),
  recent_ledger: z.array(
    z.object({
      id: z.string(),
      created_at: z.number(),
      kind: z.string(),
      delta_rial: integerText,
      balance_after_rial: integerText,
      reason: z.string(),
    })
  ),
})

export type Reseller = z.infer<typeof ResellerSchema>
export type Receipt = z.infer<typeof ReceiptSchema>
export type Instructions = z.infer<typeof InstructionsSchema>
export type BillingBook = z.infer<typeof BillingBookSchema>
export type ResellerInput = {
  interval_days: string
  grace_days: string
  recurring_fee_rial: string
  name: string
  billing_mode: 'prepaid' | 'postpaid'
  rate_rial: string | null
  interface_ids: number[]
  default_interface_id: number | null
  max_active_peers: string | null
  max_peer_traffic_bytes: string | null
  max_total_traffic_bytes: string | null
  max_duration_days: string | null
  username?: string
  password?: string
}

export const ResellerResourceSchema = ResellerSchema.shape.resources.element
export const PeerPurchasesSchema = z.array(
  z.object({
    last_error: z.string().default(''),
    next_attempt_at: z.number().default(0),
    id: z.string(),
    peer_id: z.number(),
    state: z.string(),
    amount_rial: integerText,
    attempts: z.number(),
  })
)

export const RouterMigrationsSchema = z.array(
  z.object({
    id: z.string(),
    state: z.string(),
    destination_interface_id: z.number(),
    total_peers: z.number(),
    completed_peers: z.number(),
    attempts: z.number(),
    needs_attention: z.boolean(),
  })
)

export type PeerPurchaseInput = {
  operation: string
  bytes: string
  days: string
}

export type PeerMutationInput = {
  download_bandwidth?: string | null
  upload_bandwidth?: string | null
  action: string
  value: string
  enabled: boolean | null
}

export type ResellerPeerInput = {
  download_bandwidth?: string | null
  upload_bandwidth?: string | null
  name: string
  interface_id: number
  quota_bytes: string
  duration_days: string
}

export type RouterMigrationInput = {
  interface_id: number
  reason: string
  endpoint: string
}

export type OwnershipInput = {
  reseller_id: number | null
  reason: string
  configured_duration_days: string | null
}

export type ReceiptReviewInput = {
  id: number
  action: 'approve' | 'reject'
  reason?: string
}

export type BillingActionInput = {
  action: string
  data: Record<string, unknown>
  key: string
}

// Form values stay strings until exact rial/byte conversion at submission.
export const ResellerFormSchema = z.object({
  name: z.string().trim().min(1).max(255),
  username: z.string().trim().min(1).max(64),
  password: z.string(),
  mode: z.enum(['prepaid', 'postpaid']),
  rate: z.string(),
  interval: z.string(),
  grace: z.string(),
  fee: z.string(),
  max_active: z.string(),
  max_duration: z.string(),
  max_peer: z.string(),
  max_total: z.string(),
  interface_ids: z.array(z.number()),
  default_interface: z.string(),
})

export type ResellerFormValues = z.infer<typeof ResellerFormSchema>
