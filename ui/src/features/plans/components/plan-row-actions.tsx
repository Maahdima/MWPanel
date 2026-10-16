import { useState } from 'react'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { IconEdit, IconTrash } from '@tabler/icons-react'
import { toast } from 'sonner'
import { VpnPlan } from '@/schema/vpn-plan.ts'
import { handleServerError } from '@/utils/handle-server-error.ts'
import { Button } from '@/components/ui/button.tsx'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog.tsx'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu.tsx'
import { PlanFormDialog } from '@/features/plans/components/plan-form-dialog.tsx'
import { useDeleteVpnPlanMutation } from '@/hooks/vpn-plans/useVpnPlanMutations.ts'

export function PlanRowActions({ plan }: { plan: VpnPlan }) {
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const remove = useDeleteVpnPlanMutation()

  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button variant='ghost' className='h-8 w-8 p-0'>
            <DotsHorizontalIcon className='h-4 w-4' />
            <span className='sr-only'>Open menu</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem onClick={() => setEditOpen(true)}>
            Edit
            <IconEdit size={16} className='ml-auto' />
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setDeleteOpen(true)}>
            Delete
            <IconTrash size={16} className='ml-auto' />
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <PlanFormDialog open={editOpen} plan={plan} onOpenChange={setEditOpen} />
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {plan.name}?</AlertDialogTitle>
            <AlertDialogDescription>
              Plans that already have orders stay in the list. Turn off “Show
              in the bot” to hide one instead.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={remove.isPending}>Cancel</AlertDialogCancel>
            <Button
              variant='destructive'
              disabled={remove.isPending}
              onClick={() => {
                remove.mutate(plan.id, {
                  onSuccess: () => {
                    toast.success('Plan deleted')
                    setDeleteOpen(false)
                  },
                  onError: (error) => handleServerError(error),
                })
              }}
            >
              Delete
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
