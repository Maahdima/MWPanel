'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreateInterfaceRequest,
  CreateInterfaceSchema,
  UpdateInterfaceRequest,
  UpdateInterfaceSchema,
} from '@/schema/interfaces.ts'
import { toast } from 'sonner'
import { useCreateInterfaceMutation } from '@/hooks/interfaces/useCreateInterfaceMutation.ts'
import { useUpdateInterfaceMutation } from '@/hooks/interfaces/useUpdateInterfaceMutation.ts'
import { Form } from '@/components/ui/form'
import { SimpleField } from '@/features/shared-components/simple-field.tsx'

interface Props {
  currentRow?: Partial<CreateInterfaceRequest>
  onClose: () => void
}

export function PoolForm({ currentRow, onClose }: Props) {
  const isEdit = Boolean(currentRow)
  const { mutateAsync: createInterface } = useCreateInterfaceMutation()
  const { mutateAsync: updateInterface } = useUpdateInterfaceMutation()

  const form = useForm<CreateInterfaceRequest>({
    resolver: zodResolver(
      (isEdit ? UpdateInterfaceSchema : CreateInterfaceSchema) as never
    ),
    defaultValues: currentRow ?? {},
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    }
  }, [currentRow, form])

  const onSubmit = async (
    values: CreateInterfaceRequest | UpdateInterfaceRequest
  ) => {
    if (isEdit) {
      await updateInterface(values as UpdateInterfaceRequest)
    } else {
      await createInterface(values as CreateInterfaceRequest)
    }
    form.reset()
    const toastMessage = isEdit
      ? 'Pool updated successfully.'
      : 'Pool created successfully.'
    toast.success(toastMessage, { duration: 5000 })
    onClose()
  }

  return (
    <Form {...form}>
      <form
        id='pool-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className='space-y-4'
      >
        {[
          {
            name: 'comment',
            label: 'Comment',
            placeholder: 'Comment (optional)',
          },
          {
            name: 'name',
            label: 'Name',
            placeholder: 'Server Name',
          },
          {
            name: 'ip_address',
            label: 'IP Address',
            placeholder: 'e.g., 185.51.200.10',
          },
          {
            name: 'api_port',
            label: 'API Port',
            placeholder: 'e.g., 80',
          },
          {
            name: 'username',
            label: 'Username',
            placeholder: 'Mikrotik Username',
          },
        ].map(({ name, label, placeholder }) => (
          <SimpleField
            key={name}
            name={name as keyof CreateInterfaceRequest}
            label={label}
            placeholder={placeholder}
            control={form.control}
          />
        ))}
      </form>
    </Form>
  )
}
