import { useState } from 'react'
import { IconArrowsExchange } from '@tabler/icons-react'
import { RotateCcw } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { HighlightStatsCard } from '@/features/dashboard/components/highlight-stats-card.tsx'
import { useResetTotalTrafficMutation } from '@/hooks/dashboard/useResetTotalTrafficMutation.ts'

type TrafficStatsCardProps = {
  value: string | undefined
  isLoading: boolean
}

export function TrafficStatsCard({ value, isLoading }: TrafficStatsCardProps) {
  const [open, setOpen] = useState(false)
  const { mutateAsync: resetTotalTraffic, isPending } =
    useResetTotalTrafficMutation()

  const handleReset = async () => {
    await resetTotalTraffic()
    toast.success('Total traffic usage reset successfully', {
      duration: 5000,
    })
    setOpen(false)
  }

  return (
    <>
      <HighlightStatsCard
        title='Total Traffic'
        icon={<IconArrowsExchange />}
        value={value}
        suffix='GB'
        isLoading={isLoading}
        action={
          <Button
            variant='outline'
            size='sm'
            className='shrink-0 gap-2 border-white/20 text-white hover:bg-white/10 hover:text-white'
            aria-label='Reset total traffic usage'
            disabled={isLoading || isPending}
            onClick={() => setOpen(true)}
          >
            <RotateCcw className='size-4' />
            <span className='hidden sm:inline'>Reset</span>
          </Button>
        }
      />

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader className='items-start'>
            <DialogTitle>Reset total traffic usage?</DialogTitle>
            <DialogDescription className='text-left'>
              This will reset the accumulated traffic counter to zero. Peer
              usage statistics will not be affected.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant='outline' onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleReset}
              disabled={isPending}
              className='bg-destructive dark:bg-destructive/60 hover:bg-destructive focus-visible:ring-destructive text-white'
            >
              Reset
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
