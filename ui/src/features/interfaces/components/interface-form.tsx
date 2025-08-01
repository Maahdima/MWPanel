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
  setPending: (pending: boolean) => void
}

export function InterfaceForm({ currentRow, onClose, setPending }: Props) {
  const isEdit = Boolean(currentRow)
  const { mutateAsync: createInterface, isPending: isCreateInterfacePending } =
    useCreateInterfaceMutation()
  const { mutateAsync: updateInterface, isPending: isUpdateServerPending } =
    useUpdateInterfaceMutation()

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

  useEffect(() => {
    setPending?.(isEdit ? isUpdateServerPending : isCreateInterfacePending)
  }, [isCreateInterfacePending, isUpdateServerPending, isEdit, setPending])

  const onSubmit = async (
    values: CreateInterfaceRequest | UpdateInterfaceRequest
  ) => {
    try {
      setPending(true)
      if (isEdit) {
        await updateInterface(values as UpdateInterfaceRequest)
      } else {
        await createInterface(values as CreateInterfaceRequest)
      }

      form.reset()
      toast.success(
        isEdit
          ? 'Interface updated successfully.'
          : 'Interface created successfully.',
        { duration: 5000 }
      )
      onClose()
    } finally {
      setPending(false)
    }
  }

  return (
    <Form {...form}>
      <form
        id='interface-form'
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
            placeholder: 'Interface Name',
          },
          {
            name: 'listen_port',
            label: 'API Port',
            placeholder: 'e.g., 13231',
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
