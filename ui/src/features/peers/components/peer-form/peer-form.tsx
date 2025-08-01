'use client'

import { CreatePeerRequest } from '@/schema/peers.ts'
import { Form } from '@/components/ui/form'
import { PeerFormFields } from '@/features/peers/components/peer-form/peer-form-fields.tsx'
import { PeerFormSkeleton } from '@/features/peers/components/peer-form/peer-form-skeleton.tsx'
import { usePeerForm } from '@/features/peers/components/peer-form/use-peer-form.tsx'

interface Props {
  currentRow?: Partial<CreatePeerRequest>
  onClose: () => void
  setIsLoading?: (loading: boolean) => void
}

export function PeerForm({ currentRow, onClose, setIsLoading }: Props) {
  const { form, onSubmit, isDefaultsReady } = usePeerForm({
    currentRow,
    onClose,
    setIsLoading,
  })

  if (!isDefaultsReady) {
    return <PeerFormSkeleton />
  }

  return (
    <Form {...form}>
      <form
        id='peer-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className='space-y-4'
      >
        <PeerFormFields
          control={form.control}
          setValue={form.setValue}
          isEdit={!!currentRow}
        />
      </form>
    </Form>
  )
}
