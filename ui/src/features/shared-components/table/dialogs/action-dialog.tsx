import { useState } from 'react'
import { Button } from '@/components/ui/button.tsx'
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog.tsx'
import { InterfaceForm } from '@/features/interfaces/components/interface-form.tsx'
import { PoolForm } from '@/features/ip-pools/components/pool-form.tsx'
import { PeerForm } from '@/features/peers/components/peer-form/peer-form.tsx'
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
  const [pending, setPending] = useState(false)
  const maxWidth =
    formId === 'peer-form'
      ? 'sm:max-w-xl'
      : formId === 'server-form'
        ? 'sm:max-w-lg'
        : 'sm:max-w-md'

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={maxWidth}>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        <DialogBody>
          {formId === 'server-form' ? (
            <ServerForm
              currentRow={currentRow}
              onClose={() => onOpenChange(false)}
              setPending={setPending}
            />
          ) : formId === 'interface-form' ? (
            <InterfaceForm
              currentRow={currentRow}
              onClose={() => onOpenChange(false)}
              setPending={setPending}
            />
          ) : formId === 'pool-form' ? (
            <PoolForm
              currentRow={currentRow}
              onClose={() => onOpenChange(false)}
              setPending={setPending}
            />
          ) : (
            <PeerForm
              currentRow={currentRow}
              onClose={() => onOpenChange(false)}
              setIsLoading={setPending}
            />
          )}
        </DialogBody>

        <DialogFooter>
          <Button
            variant='outline'
            type='button'
            onClick={() => onOpenChange(false)}
            disabled={pending}
          >
            Cancel
          </Button>
          <Button disabled={pending} type='submit' form={formId}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
