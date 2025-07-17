import { LoaderCircleIcon } from 'lucide-react'
import { useCreateServerMutation } from '@/hooks/servers/useCreateServerMutation.ts'
import { useUpdateServerMutation } from '@/hooks/servers/useUpdateServerMutation.ts'
import { Button } from '@/components/ui/button.tsx'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog.tsx'
import { InterfaceForm } from '@/features/interfaces/components/interface-form.tsx'
import { PeerForm } from '@/features/peers/components/peer-form.tsx'
import { ServerForm } from '@/features/servers/components/server-form.tsx'

interface Props<T> {
  currentRow?: Partial<T>
  open: boolean
  onOpenChange: (state: boolean) => void
  title: string
  description: string
  formId: string
}

export function ActionDialog<T>({
  currentRow,
  open,
  onOpenChange,
  title,
  description,
  formId,
}: Props<T>) {
  const { mutateAsync: createServer, isPending: createServerPending } =
    useCreateServerMutation()
  const { mutateAsync: updateServer, isPending: updateServerPending } =
    useUpdateServerMutation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-lg'>
        <DialogHeader className='text-left'>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        {formId === 'server-form' ? (
          <ServerForm
            createServer={async (data) => {
              await createServer(data)
            }}
            updateServer={async (data) => {
              await updateServer(data)
            }}
            currentRow={currentRow}
            onClose={() => onOpenChange(false)}
          />
        ) : formId === 'interface-form' ? (
          <InterfaceForm
            currentRow={currentRow}
            onClose={() => onOpenChange(false)}
          />
        ) : (
          <PeerForm
            currentRow={currentRow}
            onClose={() => onOpenChange(false)}
          />
        )}

        <DialogFooter>
          <Button
            disabled={createServerPending || updateServerPending}
            type='submit'
            form={formId}
          >
            {(createServerPending || updateServerPending) && (
              <LoaderCircleIcon className='animate-spin' />
            )}
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
