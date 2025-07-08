import { useDeleteInterfaceMutation } from '@/hooks/interfaces/useDeleteInterfaceMutation.ts'
import { useInterfaces } from '@/features/interfaces/context/interfaces-context.tsx'
import { ActionDialog } from '@/features/shared-components/table/dialogs/action-dialog.tsx'
import { DeleteEntityDialog } from '@/features/shared-components/table/dialogs/delete-entity-dialog.tsx'

export function InterfacesDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useInterfaces()
  const { mutateAsync } = useDeleteInterfaceMutation()

  const handleClose = (type: typeof open) => () => {
    setOpen(type)
    setTimeout(() => setCurrentRow(null), 500)
  }

  return (
    <>
      <ActionDialog
        open={open === 'add'}
        onOpenChange={() => setOpen('add')}
        title='Add New Interface'
        description='Fill out the form to create a new interface.'
        formId='interface-form'
      />

      {currentRow && (
        <>
          <ActionDialog
            open={open === 'edit'}
            onOpenChange={handleClose('edit')}
            title='Edit Interface'
            description='Update the interface below.'
            currentRow={currentRow}
            formId='interface-form'
          />

          <DeleteEntityDialog
            open={open === 'delete'}
            onOpenChange={handleClose('delete')}
            entity={{ ...currentRow, id: String(currentRow.id) }}
            entityType='Interface'
            mutationFn={async (id: string) => mutateAsync(Number(id))}
          />
        </>
      )}
    </>
  )
}
