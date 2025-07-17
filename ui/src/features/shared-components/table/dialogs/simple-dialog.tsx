import { ReactNode } from 'react'
import { IconAlertCircle } from '@tabler/icons-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog.tsx'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  actionText: string
  mutateAsync?: () => Promise<void>
  trigger?: ReactNode
}

export function SimpleDialog({
  open,
  onOpenChange,
  title,
  description,
  actionText,
  mutateAsync,
  trigger = <Button variant='outline'>Open Dialog</Button>,
}: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            <div className='flex items-center justify-start'>
              <IconAlertCircle className='mr-1' size={22} />
              {title}
            </div>
          </DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={() => (mutateAsync ? mutateAsync() : undefined)}>
            {actionText}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
