'use client'

import { useEffect, useId, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  CreateIPPoolRequest,
  CreateIPPoolSchema,
  UpdateIPPoolRequest,
  UpdateIPPoolSchema,
} from '@/schema/ip-pool.ts'
import { toast } from 'sonner'
import { withMask } from 'use-mask-input'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery.ts'
import { useCreateIPPoolMutation } from '@/hooks/ip-pool/useCreateIPPoolMutation.ts'
import { useUpdateIPPoolMutation } from '@/hooks/ip-pool/useUpdateIPPoolMutation.ts'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input.tsx'
import { InterfaceSelect } from '@/features/peers/components/fields/interface-select.tsx'
import { SimpleField } from '@/features/shared-components/simple-field.tsx'

interface Props {
  currentRow?: Partial<CreateIPPoolRequest>
  onClose: () => void
  setPending: (pending: boolean) => void
}

export function PoolForm({ currentRow, onClose, setPending }: Props) {
  const {
    data: interfacesList = [],
    isLoading: isInterfacesLoading,
    error: interfacesError,
  } = useInterfacesListQuery()

  const { mutateAsync: createIPPool, isPending: isCreateIPPoolPending } =
    useCreateIPPoolMutation()
  const { mutateAsync: updateIPPool, isPending: isUpdateIPPoolPending } =
    useUpdateIPPoolMutation()

  const isEdit = Boolean(currentRow)

  const startId = useId()
  const endId = useId()

  const startRef = useRef<HTMLInputElement>(null)
  const endRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (startRef.current) {
      withMask('ip', {
        placeholder: '_',
        inputFormat: '099.099.099.099',
        outputFormat: '099.099.099.099',
        showMaskOnHover: false,
      })(startRef.current)
    }

    if (endRef.current) {
      withMask('ip', {
        placeholder: '_',
        inputFormat: '099.099.099.099',
        outputFormat: '099.099.099.099',
        showMaskOnHover: false,
      })(endRef.current)
    }
  }, [startRef, endRef])

  const form = useForm<CreateIPPoolRequest>({
    resolver: zodResolver(
      (isEdit ? UpdateIPPoolSchema : CreateIPPoolSchema) as never
    ),
    defaultValues: currentRow ?? {},
  })

  useEffect(() => {
    if (!isEdit && interfacesList.length > 0 && !isInterfacesLoading) {
      const defaultInterface = interfacesList[0]

      form.reset((prev) => ({
        ...prev,
        interface_id: defaultInterface.id,
      }))
    }
  }, [form, isEdit, interfacesList, isInterfacesLoading])

  useEffect(() => {
    if (currentRow) {
      form.reset(currentRow)
    }
  }, [currentRow, form])

  useEffect(() => {
    setPending?.(isEdit ? isUpdateIPPoolPending : isCreateIPPoolPending)
  }, [isCreateIPPoolPending, isUpdateIPPoolPending, isEdit, setPending])

  const onSubmit = async (
    values: CreateIPPoolRequest | UpdateIPPoolRequest
  ) => {
    try {
      setPending(true)

      const patchedValues = {
        ...values,
        start_ip: values.start_ip.endsWith('/32')
          ? values.start_ip
          : `${values.start_ip}/32`,
        end_ip: values.end_ip.endsWith('/32')
          ? values.end_ip
          : `${values.end_ip}/32`,
      }

      if (isEdit) {
        await updateIPPool(patchedValues as UpdateIPPoolRequest)
      } else {
        await createIPPool(patchedValues as CreateIPPoolRequest)
      }

      form.reset()
      toast.success(
        isEdit ? 'Pool updated successfully.' : 'Pool created successfully.',
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
        id='pool-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className='space-y-4'
      >
        <SimpleField
          name='name'
          label='Name'
          placeholder='e.g., My IP Pool'
          control={form.control}
        />
        <InterfaceSelect
          name='interface_id'
          label='Interface'
          control={form.control}
          setValue={form.setValue}
          options={interfacesList}
          isLoading={isInterfacesLoading}
          error={interfacesError}
        />

        <FormField
          control={form.control}
          name='start_ip'
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor={startId}>Start IP Address</FormLabel>
              <FormControl>
                <div className='relative'>
                  <Input
                    id={startId}
                    type='text'
                    placeholder='e.g., 10.0.0.2'
                    {...field}
                    ref={(el) => {
                      field.ref(el)
                      startRef.current = el
                    }}
                    maxLength={15}
                  />
                  <div className='text-muted-foreground bg-muted border-input absolute inset-y-0 right-0 flex items-center rounded-none rounded-r-md border-l px-2 text-sm'>
                    /32
                  </div>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='end_ip'
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor={endId}>End IP Address</FormLabel>
              <FormControl>
                <div className='relative'>
                  <Input
                    id={endId}
                    type='text'
                    placeholder='e.g., 10.0.0.255'
                    {...field}
                    ref={(el) => {
                      field.ref(el)
                      endRef.current = el
                    }}
                    maxLength={15}
                  />
                  <div className='text-muted-foreground bg-muted border-input absolute inset-y-0 right-0 flex items-center rounded-none rounded-r-md border-l px-2 text-sm'>
                    /32
                  </div>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  )
}
