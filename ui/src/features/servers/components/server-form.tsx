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
import { ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import { useCreateServerMutation } from '@/hooks/servers/useCreateServerMutation.ts'
import { useUpdateServerMutation } from '@/hooks/servers/useUpdateServerMutation.ts'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form'
import { PasswordField } from '@/features/shared-components/password-field.tsx'
import { SimpleField } from '@/features/shared-components/simple-field.tsx'

type FormFieldConfig = {
  name: keyof CreateServerRequest
  label: string
  placeholder: string
}

const simpleFormFields: FormFieldConfig[] = [
  {
    name: 'name',
    label: 'Name',
    placeholder: 'Server Name',
  },
  {
    name: 'comment',
    label: 'Comment',
    placeholder: 'Comment (optional)',
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
]

interface Props {
  currentRow?: Partial<CreateServerRequest>
  onClose: () => void
  setPending: (pending: boolean) => void
}

export function ServerForm({ currentRow, onClose, setPending }: Props) {
  const isEdit = Boolean(currentRow?.name)
  const { mutateAsync: createServer, isPending: isCreateServerPending } =
    useCreateServerMutation()
  const { mutateAsync: updateServer, isPending: isUpdateServerPending } =
    useUpdateServerMutation()

  const form = useForm<CreateServerRequest>({
    resolver: zodResolver(
      isEdit ? UpdateServerSchema : (CreateServerSchema as never)
    ) as never,
    defaultValues: currentRow ?? {
      is_ssl: false,
    },
  })

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    }
  }, [currentRow, form])

  useEffect(() => {
    setPending?.(isEdit ? isUpdateServerPending : isCreateServerPending)
  }, [isCreateServerPending, isUpdateServerPending, isEdit, setPending])

  const onSubmit = async (
    values: CreateServerRequest | UpdateServerRequest
  ) => {
    try {
      setPending(true)
      if (isEdit) {
        await updateServer(values as UpdateServerRequest)
      } else {
        await createServer(values as CreateServerRequest)
      }

      form.reset()
      toast.success(
        isEdit
          ? 'Server updated successfully.'
          : 'Server created successfully.',
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
        id='server-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className='space-y-4'
      >
        <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
          {simpleFormFields.map(({ name, label, placeholder }) => (
            <SimpleField
              key={name}
              name={name}
              label={label}
              placeholder={placeholder}
              control={form.control}
            />
          ))}
        </div>

        <SimpleField
          key='username'
          name='username'
          label='Username'
          placeholder='Mikrotik Username'
          control={form.control}
        />
        <PasswordField
          name='password'
          label='Password'
          placeholder='Mikrotik Password'
          control={form.control}
        />

        <FormField
          control={form.control}
          name='is_ssl'
          render={({ field }) => (
            <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm'>
              <div className='space-y-0.5'>
                <FormLabel className='flex items-center text-base'>
                  <ShieldCheck className='mr-2 h-5 w-5 text-green-600' />
                  Use SSL/TLS
                </FormLabel>
                <FormDescription>
                  Connect to the server's API using a secure connection.
                </FormDescription>
              </div>
              <FormControl>
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
