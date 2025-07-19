import { useEffect, useState } from 'react'
import { Peer } from '@/schema/peers.ts'
import { CalendarIcon, ClipboardCopyIcon, LinkIcon, XIcon } from 'lucide-react'
import { toast } from 'sonner'
import { usePeerShareQuery } from '@/hooks/peers/usePeerShareQuery.ts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { SimpleDatepicker } from '@/features/shared-components/simple-date-picker.tsx'

type Props = {
  open: boolean
  onOpenChange: (state: boolean) => void
  currentRow: Peer
}

export function PeersShareDialog({ open, onOpenChange, currentRow }: Props) {
  const { data: shareData, isLoading } = usePeerShareQuery(currentRow.uuid, {
    enabled: open,
  })

  const [expireDate, setExpireDate] = useState<string | null>(null)

  useEffect(() => {
    if (shareData?.expire_time) {
      setExpireDate(shareData.expire_time)
    }
  }, [shareData])

  const handleCopy = () => {
    if (shareData?.share_link) {
      navigator.clipboard.writeText(shareData.share_link)
      toast.success('Share link copied to clipboard', {
        duration: 5000,
      })
    }
  }

  const handleStopSharing = () => {
    // TODO: add actual API call to stop sharing if needed
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='space-y-2 sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Share Peer</DialogTitle>
        </DialogHeader>

        <div className='gap-2 space-y-4'>
          <div className='space-y-3'>
            <Label className='flex items-center gap-1'>
              <LinkIcon className='h-4 w-4 opacity-60' />
              Share Link
            </Label>
            <div className='flex items-center gap-2'>
              <Input value={shareData?.share_link ?? ''} readOnly />
              <Button variant='outline' size='icon' onClick={handleCopy}>
                <ClipboardCopyIcon className='h-4 w-4' />
              </Button>
            </div>
          </div>

          <div className='space-y-3'>
            <Label className='flex items-center gap-1'>
              <CalendarIcon className='h-4 w-4 opacity-60' />
              Expire At
            </Label>
            <SimpleDatepicker
              value={expireDate}
              onChange={(value) => setExpireDate(value)}
              placeholder='Pick an expiration date'
            />
          </div>

          <Button
            variant='destructive'
            className='w-full'
            onClick={handleStopSharing}
          >
            <XIcon className='mr-2 h-4 w-4' />
            Stop Sharing
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
