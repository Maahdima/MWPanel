import { useState } from 'react'
import { IconAlertTriangle } from '@tabler/icons-react'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert.tsx'
import { Input } from '@/components/ui/input.tsx'
import { Label } from '@/components/ui/label.tsx'
import { ConfirmDialog } from '@/components/confirm-dialog.tsx'

interface Props<T> {
  open: boolean
  onOpenChange: (open: boolean) => void
  entity: T & { id: string; name: string }
  entityType: string
  mutationFn: (id: string) => Promise<void>
}

export function DeleteEntityDialog<T>({
  open,
  onOpenChange,
  entity,
  entityType,
  mutationFn,
}: Props<T>) {
  const [loading, setLoading] = useState(false)
  const [value, setValue] = useState('')

  const handleDelete = async () => {
    if (value.trim() !== entity.name) return
    try {
      setLoading(true)
      await mutationFn(entity.id)
      onOpenChange(false)
      toast.success(
        `${entityType} ${entity.name} has been deleted successfully`,
        { duration: 5000 }
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={value.trim() !== entity.name}
      title={
        <span className='text-destructive'>
          <IconAlertTriangle
            className='stroke-destructive mr-1 inline-block'
            size={18}
          />
          Delete {entityType}
        </span>
      }
      desc={
        <div className='space-y-4'>
          <p className='mb-2'>
            Are you sure you want to delete{' '}
            <span className='font-bold'>{entity.name}</span>?
            <br />
            This action will permanently remove the {entityType.toLowerCase()}{' '}
            with the name of <span className='font-bold'>{entity.name}</span>{' '}
            from the system. This cannot be undone.
          </p>

          <Label className='my-2'>
            Name:
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={`Enter ${entityType.toLowerCase()} name to confirm deletion.`}
            />
          </Label>

          <Alert variant='destructive'>
            <AlertTitle>Warning!</AlertTitle>
            <AlertDescription>
              Please be careful, this operation cannot be rolled back.
            </AlertDescription>
          </Alert>
        </div>
      }
      isLoading={loading}
      confirmText='Delete'
      destructive
    />
  )
}
