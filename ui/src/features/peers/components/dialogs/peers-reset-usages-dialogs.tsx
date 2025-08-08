'use client'

import { useState } from 'react'
import { RotateCcw as IconRestore, TriangleAlertIcon } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

type ResetUsagesDialogProps = {
  isPending: boolean
  resetUsages: () => Promise<void>
}

export const ResetUsagesDialog = ({
  isPending,
  resetUsages,
}: ResetUsagesDialogProps) => {
  const [open, setOpen] = useState(false)

  const handleConfirm = async () => {
    await resetUsages()
    toast.success('Peer usages reset successfully', {
      duration: 5000,
    })
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant='outline'
          className='gap-2 border-amber-500 text-amber-600 transition-all hover:bg-amber-100/60 dark:border-amber-400 dark:text-amber-400 dark:hover:bg-amber-400/10'
          disabled={isPending}
          onClick={() => setOpen(true)}
        >
          <IconRestore className='h-4 w-4' />
          <span className='text-sm font-medium'>Reset Usages</span>
        </Button>
      </DialogTrigger>

      <DialogContent>
        <DialogHeader className='items-start'>
          <div className='bg-destructive/10 mb-2 flex size-12 items-center justify-center rounded-full'>
            <TriangleAlertIcon className='size-6 text-amber-400' />
          </div>
          <DialogTitle>Reset all peer usages?</DialogTitle>
          <DialogDescription className='text-left'>
            This action cannot be undone. All peer usage data will be reset!
          </DialogDescription>
        </DialogHeader>

        <DialogFooter>
          <Button variant='outline' onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={isPending}
            className='bg-destructive dark:bg-destructive/60 hover:bg-destructive focus-visible:ring-destructive text-white'
          >
            Reset
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
