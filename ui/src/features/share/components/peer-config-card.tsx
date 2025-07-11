'use client'

import { useEffect, useState } from 'react'
import { CopyIcon } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface ConfigCardProps {
  isLoading: boolean
  blob?: Blob
  peerName?: string
}

export default function PeerConfigCard({
  isLoading,
  blob,
  peerName,
}: ConfigCardProps) {
  const [configText, setConfigText] = useState<string>('')

  useEffect(() => {
    if (blob) {
      const reader = new FileReader()
      reader.onload = () => setConfigText(reader.result as string)
      reader.readAsText(blob)
    }
  }, [blob])

  const handleCopy = async () => {
    if (!configText) return
    await navigator.clipboard.writeText(configText)
    toast.success('Copied to clipboard.', { duration: 5000 })
  }

  const handleDownload = () => {
    if (!configText) return
    const file = new Blob([configText], { type: 'text/plain;charset=utf-8' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(file)
    link.download = `${peerName || 'peer'}.conf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <Card className='gap-3'>
      <CardHeader>
        <CardTitle>Configuration</CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className='space-y-2'>
            <Skeleton className='h-4 w-full' />
            <Skeleton className='h-4 w-5/6' />
            <Skeleton className='h-4 w-4/6' />
            <Skeleton className='h-4 w-3/6' />
            <Skeleton className='mt-4 h-10 w-32' />
          </div>
        ) : (
          <>
            <div className='bg-muted relative max-h-[60vh] min-h-[10vh] overflow-auto rounded-md px-4 py-3'>
              <pre className='text-sm break-words whitespace-pre-wrap'>
                <code>{configText}</code>
              </pre>
              <Button
                variant='ghost'
                size='sm'
                onClick={handleCopy}
                className='absolute top-2 right-2'
              >
                <CopyIcon className='mr-1 h-4 w-4' />
                Copy
              </Button>
            </div>
            <Button className='mt-4' onClick={handleDownload}>
              Download Config
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  )
}
