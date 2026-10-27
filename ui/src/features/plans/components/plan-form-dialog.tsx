import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Plan, UpsertPlan, UpsertPlanSchema } from '@/schema/plan'
import { toast } from 'sonner'
import { handleServerError } from '@/utils/handle-server-error.ts'
import {
  useCreatePlanMutation,
  useUpdatePlanMutation,
} from '@/hooks/plans/usePlanMutations.ts'
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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form.tsx'
import { Input } from '@/components/ui/input.tsx'
import { Switch } from '@/components/ui/switch.tsx'

type Props = {
  open: boolean
  plan?: Plan | null
  onOpenChange: (open: boolean) => void
}

const emptyValues: UpsertPlan = {
  name: '',
  fa_name: '',
  traffic_gb: 2,
  duration_days: 30,
  enabled: true,
  sort_order: 0,
}

export function PlanFormDialog({ open, plan, onOpenChange }: Props) {
  const createPlan = useCreatePlanMutation()
  const updatePlan = useUpdatePlanMutation()
  const form = useForm<UpsertPlan>({
    resolver: zodResolver(UpsertPlanSchema) as never,
    defaultValues: emptyValues,
  })

  useEffect(() => {
    if (!open) return
    form.reset(
      plan
        ? {
            name: plan.name,
            fa_name: plan.fa_name,
            traffic_gb: plan.traffic_gb,
            duration_days: plan.duration_days,
            enabled: plan.enabled,
            sort_order: plan.sort_order,
          }
        : emptyValues
    )
  }, [open, plan, form])

  const pending = createPlan.isPending || updatePlan.isPending

  function onSubmit(values: UpsertPlan) {
    const request = plan
      ? updatePlan.mutateAsync({ id: plan.id, plan: values })
      : createPlan.mutateAsync(values)
    request
      .then(() => {
        toast.success(plan ? 'Plan updated' : 'Plan added')
        onOpenChange(false)
      })
      .catch((error) => handleServerError(error))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-xl'>
        <DialogHeader>
          <DialogTitle>{plan ? 'Edit Plan' : 'Add Plan'}</DialogTitle>
          <DialogDescription>
            {plan
              ? 'Update the plan below.'
              : 'Fill out the form to add a plan. The panel shows the name. The bot shows the Persian name.'}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <Form {...form}>
            <form
              id='plan-form'
              className='grid grid-cols-1 gap-x-4 gap-y-4 @min-[28rem]:grid-cols-2'
              onSubmit={form.handleSubmit(onSubmit)}
            >
              <div className='@min-[28rem]:col-span-2'>
                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl>
                        <Input placeholder='2 GB / 30 days' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div className='@min-[28rem]:col-span-2'>
                <FormField
                  control={form.control}
                  name='fa_name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Persian name</FormLabel>
                      <FormControl>
                        <Input
                          className='font-iranyekan'
                          dir='auto'
                          placeholder='۲ گیگ، ۳۰ روز'
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <FormField
                control={form.control}
                name='traffic_gb'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Traffic (GB)</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='duration_days'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Duration (days)</FormLabel>
                    <FormControl>
                      <Input type='number' min={1} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='sort_order'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Sort order</FormLabel>
                    <FormControl>
                      <Input type='number' min={0} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name='enabled'
                render={({ field }) => (
                  <FormItem className='flex h-full flex-row items-center justify-between gap-3 pt-6'>
                    <FormLabel>Show in the bot</FormLabel>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </form>
          </Form>
        </DialogBody>
        <DialogFooter>
          <Button
            variant='outline'
            type='button'
            disabled={pending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type='submit' form='plan-form' disabled={pending}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
