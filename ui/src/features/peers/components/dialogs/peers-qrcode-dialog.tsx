import { Peer } from '@/schema/peers.ts'
import { usePeerQRCodeQuery } from '@/hooks/peers/usePeerQRCodeQuery.ts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'

interface Props {
  open: boolean
  onOpenChange: (state: boolean) => void
  currentRow: Peer
}

export function PeersQRCodeDialog({ open, onOpenChange, currentRow }: Props) {
  const { data: peerQRCode, isLoading } = usePeerQRCodeQuery(currentRow.uuid, {
    enabled: open,
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader className='text-left'>
          <DialogTitle>QR Code</DialogTitle>
          <DialogDescription>
            Scan this code with a compatible device.
          </DialogDescription>
        </DialogHeader>

        <div className='flex justify-center py-6'>
          {isLoading ? (
            <Skeleton className='h-48 w-48 rounded-md' />
          ) : peerQRCode ? (
            <img
              src={peerQRCode}
              alt='Peer QR Code'
              className='rounded-md border object-contain'
            />
          ) : (
            <div className='text-muted-foreground'>QR Code not available</div>
          )}
        </div>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
