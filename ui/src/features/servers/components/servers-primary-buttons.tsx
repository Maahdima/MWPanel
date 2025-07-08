import { IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useServers } from '@/features/servers/context/servers-context.tsx'

interface Props {
  refetchServersList: () => void
  isServersListRefetching: boolean
}

export function ServersPrimaryButtons({
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
        <span>{isServersListRefetching ? 'Please wait' : 'Refresh'}</span>{' '}
        {isServersListRefetching ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconRefresh />
        )}
      </Button>
      <Button
        className='space-x-1'
        disabled={true}
        onClick={() => setOpen('add')}
      >
        <span>Add New Server</span> <IconUserPlus size={18} />
      </Button>
    </div>
  )
}
