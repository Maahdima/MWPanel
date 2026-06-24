import { useState } from 'react'
import { IconArrowsExchange } from '@tabler/icons-react'
import { RotateCcw } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
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

  const displayValue = value ? `${value} GB` : '-'

  return (
    <>
      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <CardTitle className='text-lg font-medium'>Total Traffic</CardTitle>
          <IconArrowsExchange />
        </CardHeader>
        <CardContent>
          <div className='flex items-center justify-between gap-2'>
            {isLoading ? (
              <Skeleton className='h-6 w-16 rounded-sm' />
            ) : (
              <div className='text-2xl font-bold'>{displayValue}</div>
            )}
            <Button
              variant='ghost'
              size='icon'
              className='text-muted-foreground hover:text-foreground h-8 w-8 shrink-0'
              aria-label='Reset total traffic usage'
              disabled={isLoading || isPending}
              onClick={() => setOpen(true)}
            >
              <RotateCcw className='h-4 w-4' />
            </Button>
          </div>
        </CardContent>
      </Card>

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
