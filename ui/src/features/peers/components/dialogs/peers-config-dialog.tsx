import { useEffect, useRef, useState } from 'react'
import { Peer } from '@/schema/peer'
import { CopyIcon } from 'lucide-react'
import { toast } from 'sonner'
import { downloadWireGuardConfig } from '@/utils/helper.ts'
import { usePeerConfigQuery } from '@/hooks/peers/usePeerConfigQuery.ts'
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
  download?: boolean
}

export function PeersConfigDialog({
  open,
  onOpenChange,
  currentRow,
  download = false,
}: Props) {
  const { data: peerConfigBlob, isLoading } = usePeerConfigQuery(
    currentRow.id,
    {
      enabled: open,
    }
  )
  const [configText, setConfigText] = useState<string>('')
  const hasDownloaded = useRef(false)

  useEffect(() => {
    if (peerConfigBlob) {
      const reader = new FileReader()
      reader.onload = () => {
        const result = reader.result as string
        setConfigText(result)
      }
      reader.readAsText(new Blob([peerConfigBlob]))

      if (download && !hasDownloaded.current) {
        hasDownloaded.current = true
        downloadWireGuardConfig(peerConfigBlob, currentRow.name)
        onOpenChange(false)
      }
    }
  }, [peerConfigBlob, download, currentRow.name, onOpenChange])

  const handleCopy = async () => {
    if (!configText) return
    await navigator.clipboard.writeText(configText)
    toast.success('Copied to clipboard.', { duration: 5000 })
  }

  if (download) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>Configuration</DialogTitle>
          <DialogDescription>
            Copy this WireGuard config or download it from the peer actions.
          </DialogDescription>
        </DialogHeader>

        <DialogBody>
          <div className='bg-muted relative overflow-auto rounded-xl px-3.5 py-3'>
            {isLoading ? (
              <Skeleton className='h-40 w-full' />
            ) : configText ? (
              <pre className='text-[13px] leading-relaxed break-words whitespace-pre-wrap'>
                <code>{configText}</code>
              </pre>
            ) : (
              <div className='text-muted-foreground text-sm'>
                Configuration not available
              </div>
            )}

            {!isLoading && configText && (
              <Button
                variant='outline'
                size='sm'
                onClick={handleCopy}
                className='absolute top-2 right-2'
              >
                <CopyIcon className='size-3.5' />
                Copy
              </Button>
            )}
          </div>
        </DialogBody>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
