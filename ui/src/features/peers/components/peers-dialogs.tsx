import { useDeletePeerMutation } from '@/hooks/peers/useDeletePeerMutation.ts'
import { PeersConfigDialog } from '@/features/peers/components/dialogs/peers-config-dialog.tsx'
import { PeersQRCodeDialog } from '@/features/peers/components/dialogs/peers-qrcode-dialog.tsx'
import { PeersShareDialog } from '@/features/peers/components/dialogs/peers-share-dialog.tsx'
import { usePeers } from '@/features/peers/context/peers-context.tsx'
import { ActionDialog } from '@/features/shared-components/table/dialogs/action-dialog.tsx'
import { DeleteEntityDialog } from '@/features/shared-components/table/dialogs/delete-entity-dialog.tsx'

export function PeersDialogs() {
  const { open, setOpen, currentRow, setCurrentRow } = usePeers()
  const { mutateAsync } = useDeletePeerMutation()

  const handleClose = (type: typeof open) => () => {
    setOpen(type)
    setTimeout(() => setCurrentRow(null), 500)
  }

  return (
    <>
      <ActionDialog
        open={open === 'add'}
        onOpenChange={() => setOpen('add')}
        title='Create New Peer'
        description='Fill out the form to create a new peer.'
        formId='peer-form'
      />

      {currentRow && (
        <>
          <ActionDialog
            open={open === 'edit'}
            onOpenChange={handleClose('edit')}
            title='Edit Peer'
            description='Update the peer below.'
            currentRow={currentRow}
            formId='peer-form'
          />

          <DeleteEntityDialog
            open={open === 'delete'}
            onOpenChange={handleClose('delete')}
            entity={{ ...currentRow, id: String(currentRow.id) }}
            entityType='Peer'
            mutationFn={async (id: string) => mutateAsync(Number(id))}
          />

          <PeersShareDialog
            key={`peer-share-${currentRow.id}`}
            open={open === 'share'}
            onOpenChange={handleClose('share')}
            currentRow={currentRow}
          />

          <PeersQRCodeDialog
            key={`peer-qrcode-${currentRow.id}`}
            open={open === 'qrCode'}
            onOpenChange={handleClose('qrCode')}
            currentRow={currentRow}
          />

          <PeersConfigDialog
            key={`peer-config-${currentRow.id}`}
            open={open === 'show_config'}
            download={open === 'download_config'}
            onOpenChange={handleClose('config')}
            currentRow={currentRow}
          />
        </>
      )}
    </>
  )
}
