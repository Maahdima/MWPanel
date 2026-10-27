import { Peer } from '@/schema/peer'
import { usePeerQRCodeQuery } from '@/hooks/peers/usePeerQRCodeQuery.ts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogBody,
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
  const { data: peerQRCode, isLoading } = usePeerQRCodeQuery(currentRow.id, {
    enabled: open,
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-sm'>
        <DialogHeader>
          <DialogTitle>QR Code</DialogTitle>
          <DialogDescription>
            Scan this code with a compatible WireGuard client.
          </DialogDescription>
        </DialogHeader>

        <DialogBody className='flex justify-center py-3'>
          {isLoading ? (
            <Skeleton className='size-52 rounded-xl' />
          ) : peerQRCode ? (
            <img
              src={peerQRCode}
              alt='Peer QR Code'
              className='size-52 rounded-xl border bg-white object-contain p-2'
            />
          ) : (
            <div className='text-muted-foreground py-8 text-sm'>
              QR Code not available
            </div>
          )}
        </DialogBody>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
