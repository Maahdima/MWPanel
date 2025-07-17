'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreateServerRequest,
  CreateServerSchema,
  UpdateServerRequest,
  UpdateServerSchema,
} from '@/schema/servers.ts'
import { toast } from 'sonner'
import { Form } from '@/components/ui/form'
import { PasswordField } from '@/features/shared-components/password-field.tsx'
import { SimpleField } from '@/features/shared-components/simple-field.tsx'

interface ServerFormProps {
  createServer: (data: CreateServerRequest) => Promise<void>
  updateServer: (data: UpdateServerRequest) => Promise<void>
  currentRow?: Partial<CreateServerRequest>
  onClose: () => void
}

export function ServerForm({
  createServer,
  updateServer,
  currentRow,
  onClose,
}: ServerFormProps) {
  const isEdit = Boolean(currentRow)

  const form = useForm<CreateServerRequest>({
    resolver: zodResolver(
      (isEdit ? UpdateServerSchema : CreateServerSchema) as never
    ),
    defaultValues: currentRow ?? {},
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    }
  }, [currentRow, form])

  const onSubmit = async (
    values: CreateServerRequest | UpdateServerRequest
  ) => {
    if (isEdit) {
      await updateServer(values as UpdateServerRequest)
    } else {
      await createServer(values as CreateServerRequest)
    }
    form.reset()
    const toastMessage = isEdit
      ? 'Server updated successfully.'
      : 'Server created successfully.'
    toast.success(toastMessage, { duration: 5000 })
    onClose()
  }

  return (
    <Form {...form}>
      <form
        id='server-form'
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
            name={name as keyof CreateServerRequest}
            label={label}
            placeholder={placeholder}
            control={form.control}
          />
        ))}
        <PasswordField
          name='password'
          label='Password'
          placeholder='Mikrotik Password'
          control={form.control}
        />
      </form>
    </Form>
  )
}
