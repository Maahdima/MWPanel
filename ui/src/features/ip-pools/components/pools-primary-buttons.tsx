import { IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { usePools } from '@/features/ip-pools/context/pools-context.tsx'

interface Props {
  refetchIPPoolsList: () => void
  isIPPoolsListRefetching: boolean
}

export function PoolsPrimaryButtons({
  refetchIPPoolsList,
  isIPPoolsListRefetching,
}: Props) {
  const { setOpen } = usePools()

  return (
    <div className='flex gap-2'>
      <Button
        variant='outline'
        className='space-x-1'
        disabled={isIPPoolsListRefetching}
        onClick={refetchIPPoolsList}
      >
        <span>Refresh</span>
        {isIPPoolsListRefetching ? (
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
