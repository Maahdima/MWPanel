import { IconCloudDown, IconRefresh, IconUserPlus } from '@tabler/icons-react'
import { FileSpreadsheet, Loader2Icon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useResetUsagesMutation } from '@/hooks/peers/useResetUsagesMutation.ts'
import { useTrafficExcelMutation } from '@/hooks/peers/useTrafficExcelMutation.ts'
import { Button } from '@/components/ui/button'
import { ResetUsagesDialog } from '@/features/peers/components/dialogs/peers-reset-usages-dialogs.tsx'
import { usePeers } from '@/features/peers/context/peers-context.tsx'

interface Props {
  refetchPeersList: () => void
  isPeersListRefetching: boolean
}

export function PeersPrimaryButtons({
  refetchPeersList,
  isPeersListRefetching,
}: Props) {
  const { setOpen } = usePeers()

  const { mutateAsync: resetUsages, isPending: isResetUsagesPending } =
    useResetUsagesMutation()

  const {
    mutateAsync: exportTrafficExcel,
    isPending: isExportingTrafficExcel,
  } = useTrafficExcelMutation()

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <div className='inline-flex w-fit -space-x-px rounded-md shadow-xs rtl:space-x-reverse'>
        <Button
          variant='outline'
          className={cn('rounded-none rounded-s-md shadow-none transition-all focus-visible:z-10')}
          onClick={() => setOpen('sync')}
        >
          <IconCloudDown className='h-4 w-4' />
          <span className='text-sm font-medium'>Sync</span>
        </Button>

        <Button
          variant='outline'
          className={cn(
            'rounded-none rounded-e-md shadow-none focus-visible:z-10',
            isPeersListRefetching && 'cursor-not-allowed opacity-70'
          )}
          disabled={isPeersListRefetching}
          onClick={refetchPeersList}
        >
          {isPeersListRefetching ? (
            <Loader2Icon className='h-4 w-4 animate-spin' />
          ) : (
            <IconRefresh className='h-4 w-4' />
          )}
          <span className='text-sm font-medium'>Refresh</span>
        </Button>
      </div>

      <ResetUsagesDialog
        isPending={isResetUsagesPending}
        resetUsages={resetUsages}
      />

      <Button
        variant='outline'
        className='gap-2 border-b-emerald-900 text-emerald-600 transition-all hover:bg-emerald-100/60 dark:border-emerald-600 dark:text-emerald-400 dark:hover:bg-emerald-400/10'
        disabled={isExportingTrafficExcel}
        onClick={() => exportTrafficExcel()}
      >
        <span className='text-sm font-medium'>Traffic Excel</span>
        {isExportingTrafficExcel ? (
          <Loader2Icon className='h-4 w-4 animate-spin' />
        ) : (
          <FileSpreadsheet className='h-4 w-4' />
        )}
      </Button>

      <Button onClick={() => setOpen('add')} className='gap-2 transition-all'>
        <span className='text-sm font-medium'>Add Peer</span>
        <IconUserPlus className='h-4 w-4' />
      </Button>
    </div>
  )
}
