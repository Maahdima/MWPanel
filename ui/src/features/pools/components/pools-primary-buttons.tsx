import { IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { usePools } from '@/features/pools/context/pools-context.tsx'

interface Props {
  refetchServersList: () => void
  isServersListRefetching: boolean
}

export function PoolsPrimaryButtons({
  refetchServersList,
  isServersListRefetching,
}: Props) {
  const { setOpen } = usePools()

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='space-x-1'
        disabled={isServersListRefetching}
        onClick={refetchServersList}
      >
        <span>Refresh</span>
        {isServersListRefetching ? (
          <Loader2Icon className='animate-spin' />
        ) : (
          <IconRefresh />
        )}
      </Button>
      <Button className='space-x-1' onClick={() => setOpen('add')}>
        <span>Add New Pool</span> <IconUserPlus size={18} />
      </Button>
    </div>
  )
}
