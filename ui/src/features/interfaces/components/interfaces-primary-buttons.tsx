import { IconCloudDown, IconRefresh, IconWorldPlus } from '@tabler/icons-react'
import { Loader2Icon } from 'lucide-react'
import { cn } from '@/lib/utils.ts'
import { Button } from '@/components/ui/button'
import { useInterfaces } from '@/features/interfaces/context/interfaces-context.tsx'

interface Props {
  refetchInterfacesList: () => void
  isInterfacesListRefetching: boolean
  syncInterfaces: () => void
  isInterfacesSyncing: boolean
}

export function InterfacesPrimaryButtons({
  refetchInterfacesList,
  isInterfacesListRefetching,
  syncInterfaces,
  isInterfacesSyncing,
}: Props) {
  const { setOpen } = useInterfaces()

  return (
    <div className='flex gap-2'>
      <div className='inline-flex w-fit -space-x-px rounded-md shadow-xs rtl:space-x-reverse'>
        <Button
          variant='outline'
          className={cn(
            'rounded-none rounded-s-md shadow-none transition-all focus-visible:z-10',
            isInterfacesSyncing && 'cursor-not-allowed opacity-70'
          )}
          disabled={isInterfacesSyncing}
          onClick={syncInterfaces}
        >
          {isInterfacesSyncing ? (
            <Loader2Icon className='h-4 w-4 animate-spin' />
          ) : (
            <IconCloudDown className='h-4 w-4' />
          )}
          <span className='text-sm font-medium'>Sync</span>
        </Button>

        <Button
          variant='outline'
          className={cn(
            'rounded-none rounded-e-md shadow-none focus-visible:z-10',
            isInterfacesListRefetching && 'cursor-not-allowed opacity-70'
          )}
          disabled={isInterfacesListRefetching}
          onClick={refetchInterfacesList}
        >
          {isInterfacesListRefetching ? (
            <Loader2Icon className='h-4 w-4 animate-spin' />
          ) : (
            <IconRefresh className='h-4 w-4' />
          )}
          <span className='text-sm font-medium'>Refresh</span>
        </Button>
      </div>

      <Button className='space-x-1' onClick={() => setOpen('add')}>
        <span>Add Interface</span> <IconWorldPlus size={18} />
      </Button>
    </div>
  )
}
