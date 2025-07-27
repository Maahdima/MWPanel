'use client'

import { useEffect, useId, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { withMask } from 'use-mask-input'
import { useInterfacesListQuery } from '@/hooks/interfaces/useInterfacesListQuery.ts'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { InterfaceSelect } from '@/features/peers/components/fields/interface-select.tsx'

type Props = {
  open: boolean
  onOpenChange: (state: boolean) => void
  onSave?: (data: FormValues) => void
  isLoading?: boolean
  error?: unknown
}

type FormValues = {
  interface: string
  interface_id: string
  listen_port: string
  start_ip: string
  end_ip: string
}

export function PeersIpPoolDialog({
  open,
  onOpenChange,
  onSave,
  isLoading,
  error,
}: Props) {
  const {
    data: interfacesList = [],
    isLoading: isInterfacesLoading,
    error: interfacesError,
  } = useInterfacesListQuery()

  const startId = useId()
  const endId = useId()

  const form = useForm<FormValues>({})

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

  const handleSubmit = (values: FormValues) => {
    onSave?.(values)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        onOpenAutoFocus={(e) => e.preventDefault()}
        className='space-y-2 sm:max-w-md'
      >
        <DialogHeader>
          <DialogTitle>IP Pool Range</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className='space-y-4'
          >
            <InterfaceSelect
              name='interface'
              label='Select Interface'
              control={form.control}
              setValue={form.setValue}
              options={interfacesList}
              isLoading={isLoading}
              error={error}
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
                        placeholder='10.0.0.2'
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
                        placeholder='10.0.0.255'
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

            <DialogFooter>
              <Button type='submit'>Save</Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
