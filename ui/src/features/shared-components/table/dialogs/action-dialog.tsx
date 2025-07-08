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
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-lg'>
        <DialogHeader className='text-left'>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        {formId === 'server-form' ? (
          <ServerForm
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
          <Button type='submit' form={formId}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
