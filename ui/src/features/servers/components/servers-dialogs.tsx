import { useDeleteServerMutation } from '@/hooks/servers/useDeleteServerMutation.ts'
import { useServers } from '@/features/servers/context/servers-context.tsx'
import { ActionDialog } from '@/features/shared-components/table/dialogs/action-dialog.tsx'
import { DeleteEntityDialog } from '@/features/shared-components/table/dialogs/delete-entity-dialog.tsx'

export function ServersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = useServers()
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
        title='Add New Server'
        description='Fill out the form to create a new server.'
        formId='server-form'
      />

      {currentRow && (
        <>
          <ActionDialog
            open={open === 'edit'}
            onOpenChange={handleClose('edit')}
            title='Edit Server'
            description='Update the server below.'
            currentRow={currentRow}
            formId='server-form'
          />

          <DeleteEntityDialog
            open={open === 'delete'}
            onOpenChange={handleClose('delete')}
            entity={{ ...currentRow, id: String(currentRow.id) }}
            entityType='Server'
            mutationFn={async (id: string) => mutateAsync(Number(id))}
          />
        </>
      )}
    </>
  )
}
