import { useEffect, useRef, useState } from 'react'
import { Peer } from '@/schema/peers.ts'
import { CopyIcon } from 'lucide-react'
import { toast } from 'sonner'
import { usePeerConfigQuery } from '@/hooks/peers/usePeerConfigQuery.ts'
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
  download?: boolean
}

export function PeersConfigDialog({
  open,
  onOpenChange,
  currentRow,
  download = false,
}: Props) {
  const { data: peerConfigBlob, isLoading } = usePeerConfigQuery(currentRow.id)
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
        const url = URL.createObjectURL(new Blob([peerConfigBlob]))
        const a = document.createElement('a')
        a.href = url
        a.download = `${currentRow.name || 'peer'}.conf`
        a.click()
        URL.revokeObjectURL(url)

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
      <DialogContent className='sm:max-w-xl'>
        <DialogHeader className='text-left'>
          <DialogTitle>Configuration</DialogTitle>
          <DialogDescription>Copy the configuration below.</DialogDescription>
        </DialogHeader>

        <div className='bg-muted relative max-h-[60vh] overflow-auto rounded-md px-4 py-3'>
          {isLoading ? (
            <Skeleton className='h-[200px] w-full' />
          ) : (
            <pre className='text-sm break-words whitespace-pre-wrap'>
              <code>{configText}</code>
            </pre>
          )}

          {!isLoading && (
            <Button
              variant='ghost'
              size='sm'
              onClick={handleCopy}
              className='absolute top-2 right-2'
            >
              <CopyIcon className='mr-1 h-4 w-4' />
              Copy
            </Button>
          )}
        </div>

        <DialogFooter>
          <Button onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
