import { toast } from 'sonner'
import { Switch } from '@/components/ui/switch'
import {
  useShopSalesQuery,
  useUpdateShopSalesMutation,
} from '@/hooks/shop/useShopSales.ts'

export function BotSales() {
  const { data, isLoading } = useShopSalesQuery()
  const update = useUpdateShopSalesMutation()
  const allowNew = data?.allow_new ?? true
  const allowRenew = data?.allow_renew ?? true

  const save = (next: { allow_new: boolean; allow_renew: boolean }) => {
    update.mutate(next, {
      onError: () => toast.error('Failed to update bot sales'),
    })
  }

  return (
    <div className='mb-4 rounded-xl border px-4 py-3'>
      <div className='mb-3'>
        <div className='text-sm font-medium'>Bot sales</div>
        <p className='text-muted-foreground text-sm'>
          Close new configs or renewals in the Telegram bot when capacity is
          full. A payment that is already pending still completes.
        </p>
      </div>
      <div className='flex flex-wrap gap-x-8 gap-y-3'>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={allowNew}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ allow_new: checked, allow_renew: allowRenew })
            }
          />
          New configs
        </label>
        <label className='flex items-center gap-3 text-sm'>
          <Switch
            checked={allowRenew}
            disabled={isLoading || update.isPending}
            onCheckedChange={(checked) =>
              save({ allow_new: allowNew, allow_renew: checked })
            }
          />
          Renewals
        </label>
      </div>
    </div>
  )
}
