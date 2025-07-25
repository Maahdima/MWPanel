import { Peer } from '@/schema/peers.ts'
import {
  CalendarIcon,
  CheckIcon,
  ClipboardCopyIcon,
  LinkIcon,
  XIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { usePeerShareQuery } from '@/hooks/peers/usePeerShareQuery.ts'
import { useUpdatePeerShareExpireMutation } from '@/hooks/peers/useUpdatePeerShareExpireMutation.ts'
import { useUpdatePeerShareStatusMutation } from '@/hooks/peers/useUpdatePeerShareStatusMutation.ts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { SimpleDatepicker } from '@/features/shared-components/simple-date-picker.tsx'

type Props = {
  open: boolean
  onOpenChange: (state: boolean) => void
  currentRow: Peer
}

export function PeersShareDialog({ open, onOpenChange, currentRow }: Props) {
  const {
    data: shareData,
    isLoading,
    isFetching,
  } = usePeerShareQuery(currentRow.id, {
    enabled: open,
  })

  const updatePeerShareStatus = useUpdatePeerShareStatusMutation()
  const updatePeerShareExpire = useUpdatePeerShareExpireMutation()

  const isMutating =
    updatePeerShareStatus.isPending || updatePeerShareExpire.isPending

  const handleCopy = () => {
    if (shareData?.share_link) {
      navigator.clipboard.writeText(shareData.share_link)
      toast.success('Share link copied to clipboard', { duration: 5000 })
    }
  }

  const handleExpireDateChange = async (value: string | null) => {
    await updatePeerShareExpire.mutateAsync({
      id: currentRow.id,
      expire_time: value,
    })

    toast.success(
      value
        ? 'Expiration date updated successfully'
        : 'Expiration date cleared successfully',
      { duration: 5000 }
    )
  }

  const handleToggleStatus = async () => {
    await updatePeerShareStatus.mutateAsync(currentRow.id)

    toast.success(
      shareData?.is_shared
        ? 'Peer sharing stopped successfully'
        : 'Peer sharing started successfully',
      { duration: 5000 }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='space-y-2 sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Share Peer</DialogTitle>
        </DialogHeader>

        <div className='gap-2 space-y-4'>
          {isLoading || isFetching ? (
            <>
              <Skeleton className='h-4 w-3/4' />
              <Skeleton className='h-10 w-full' />
              <Skeleton className='h-4 w-1/2' />
              <Skeleton className='h-10 w-full' />
              <Skeleton className='h-10 w-full' />
            </>
          ) : (
            <>
              {!shareData?.is_shared && (
                <Label>
                  Currently, this peer is not shared. You can start sharing it
                  by clicking the button below.
                </Label>
              )}

              {shareData?.is_shared && (
                <div className='space-y-3'>
                  <Label className='flex items-center gap-1'>
                    <LinkIcon className='h-4 w-4 opacity-60' />
                    Share Link
                  </Label>
                  <div className='flex items-center gap-2'>
                    <Input value={shareData.share_link ?? ''} readOnly />
                    <Button
                      variant='outline'
                      size='icon'
                      onClick={handleCopy}
                      disabled={isMutating}
                    >
                      <ClipboardCopyIcon className='h-4 w-4' />
                    </Button>
                  </div>
                </div>
              )}

              {shareData?.is_shared && (
                <div className='space-y-3'>
                  <Label className='flex items-center gap-1'>
                    <CalendarIcon className='h-4 w-4 opacity-60' />
                    Expire At
                  </Label>
                  <SimpleDatepicker
                    value={shareData.expire_time ?? null}
                    onChange={handleExpireDateChange}
                    placeholder='Pick an expiration date'
                  />
                </div>
              )}

              {shareData?.is_shared ? (
                <Button
                  variant='destructive'
                  className='w-full'
                  onClick={handleToggleStatus}
                  disabled={isMutating}
                >
                  <XIcon className='mr-2 h-4 w-4' />
                  Stop Sharing
                </Button>
              ) : (
                <Button
                  className='w-full text-white shadow-xs hover:bg-green-600/90 focus-visible:ring-green-600/20 dark:bg-green-600/60 dark:focus-visible:ring-green-600/40'
                  onClick={handleToggleStatus}
                  disabled={isMutating}
                >
                  <CheckIcon className='mr-2 h-4 w-4' />
                  Start Sharing
                </Button>
              )}
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
