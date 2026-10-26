import { z } from 'zod'

export const integerText = z.string().regex(/^-?\d+$/)
export const PrincipalSchema = z.object({
  account_id: z.number(),
  username: z.string(),
  role: z.enum(['superadmin', 'reseller']),
  reseller_id: z.number().optional(),
})
export const ResellerSchema = z.object({
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
