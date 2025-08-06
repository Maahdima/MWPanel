import { IconCloudPlus, IconRefresh } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useServers } from '@/features/servers/context/servers-context.tsx'

interface Props {
  serversLength: number
  refetchServersList: () => void
  isServersListRefetching: boolean
}

export function ServersPrimaryButtons({
  serversLength,
  refetchServersList,
  isServersListRefetching,
}: Props) {
  const { setOpen } = useServers()

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='space-x-1'
        disabled={isServersListRefetching}
        onClick={refetchServersList}
      >
        {isServersListRefetching ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconRefresh />
        )}
        <span>Refresh</span>
      </Button>
      <Button
        className='space-x-1'
        disabled={serversLength === 1}
        onClick={() => setOpen('add')}
      >
        <span>Add Server</span> <IconCloudPlus size={18} />
      </Button>
    </div>
  )
}
