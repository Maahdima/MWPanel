import { useDeleteServerMutation } from '@/hooks/servers/useDeleteServerMutation.ts'
import { usePools } from '@/features/pools/context/pools-context.tsx'
import { ActionDialog } from '@/features/shared-components/table/dialogs/action-dialog.tsx'
import { DeleteEntityDialog } from '@/features/shared-components/table/dialogs/delete-entity-dialog.tsx'

export function PoolsDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = usePools()
  const { mutateAsync } = useDeleteServerMutation()

  const handleClose = (type: typeof open) => () => {
    setOpen(type)
    setTimeout(() => setCurrentRow(null), 500)
  }

  return (
    <>
      <ActionDialog
        open={open === 'add'}
        onOpenChange={() => setOpen('add')}
        title='Add New Pool'
        description='Fill out the form to create a new pool.'
        formId='pool-form'
      />

      {currentRow && (
        <>
          <ActionDialog
            open={open === 'edit'}
            onOpenChange={handleClose('edit')}
            title='Edit Pool'
            description='Update the pool below.'
            currentRow={currentRow}
            formId='pool-form'
          />

          <DeleteEntityDialog
            open={open === 'delete'}
            onOpenChange={handleClose('delete')}
            entity={{ ...currentRow, id: String(currentRow.id) }}
            entityType='Pool'
            mutationFn={async (id: string) => mutateAsync(Number(id))}
          />
        </>
      )}
    </>
  )
}
