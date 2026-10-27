import type { ShopSales } from '@/schema/shop-sale'
import { toast } from 'sonner'
import {
  useShopSalesQuery,
  useUpdateShopSalesMutation,
} from '@/hooks/shop/useShopSales.ts'
import { Switch } from '@/components/ui/switch'

export function BotSales() {
  const { data, isLoading } = useShopSalesQuery()
  const update = useUpdateShopSalesMutation()
  const allowNew = data?.allow_new ?? true
  const allowRenew = data?.allow_renew ?? true
  const allowReferral = data?.allow_referral ?? false
  const requireReferral = data?.require_referral ?? false

  const save = (next: ShopSales) => {
    update.mutate(next, {
      onError: () => toast.error('Failed to update bot sales'),
    })
  }

  const current = (): ShopSales => ({
    allow_new: allowNew,
    allow_renew: allowRenew,
    allow_referral: allowReferral,
    require_referral: requireReferral,
  })

  return (
    <div className='mb-4 rounded-xl border px-4 py-3'>
      <div className='mb-3'>
        <div className='text-sm font-medium'>Bot sales</div>
        <p className='text-muted-foreground text-sm'>
          Close new configs, renewals, or referral gifts in the Telegram bot. A
          payment that is already pending still completes. The gift size is
          REFERRAL_BONUS_GB on the shop. Invite only applies to a first new
          config, never to renewals. A customer who already bought can buy
          another config without a code.
        </p>
      </div>
      <div className='flex flex-wrap items-center gap-x-8 gap-y-3'>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={allowNew}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ ...current(), allow_new: checked })
            }
          />
          New configs
        </label>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={allowRenew}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ ...current(), allow_renew: checked })
            }
          />
          Renewals
        </label>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={allowReferral}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ ...current(), allow_referral: checked })
            }
          />
          Referrals
        </label>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={requireReferral}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ ...current(), require_referral: checked })
            }
          />
          Invite only (new configs)
        </label>
      </div>
    </div>
  )
}
