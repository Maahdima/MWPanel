import { ColumnDef } from '@tanstack/react-table'
import { Interface } from '@/schema/interfaces.ts'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useUpdateInterfaceStatusMutation } from '@/hooks/interfaces/useUpdateInterfaceStatusMutation.ts'
import { Badge } from '@/components/ui/badge.tsx'
import { Switch } from '@/components/ui/switch.tsx'
import LongText from '@/components/long-text'
import { InterfacesTableRowActions } from '@/features/interfaces/components/interfaces-table-row-actions.tsx'
import { DataTableColumnHeader } from '@/features/shared-components/table/data-table-column-header.tsx'

export const interfacesColumns: ColumnDef<Interface>[] = [
  {
    id: 'is_active',
    cell: ({ row }) => {
      const iface = row.original
      const updateInterfaceStatusMutation = useUpdateInterfaceStatusMutation()

      const handleToggle = () => {
        updateInterfaceStatusMutation.mutate(iface.id, {
          onSuccess: () => {
            toast.success(
              `Interface ${!iface.disabled ? 'disabled' : 'enabled'} successfully`,
              {
                duration: 5000,
              }
            )
          },
        })
      }

      return (
        <div className='flex items-center justify-center'>
          <Switch
            id={`status-${iface.id}`}
            checked={!iface.disabled}
            onCheckedChange={handleToggle}
            disabled={updateInterfaceStatusMutation.isPending}
          />
        </div>
      )
    },
    enableSorting: false,
  },
  {
    accessorKey: 'comment',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Comment' />
    ),
    cell: ({ row }) => (
      <LongText className='max-w-36'>
        {row.getValue('comment') ?? (
          <span className='text-muted-foreground'>N/A</span>
        )}
      </LongText>
    ),
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
  },
  {
    id: 'name',
    accessorKey: 'name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Name' />
    ),
    cell: ({ row }) => {
      const { name } = row.original
      return <div className='w-fit text-nowrap'>{name}</div>
    },
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    accessorKey: 'listen_port',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Listen Port' />
    ),
    cell: ({ row }) => <div>{row.getValue('listen_port')}</div>,
    meta: {
      className: cn('border-l border-r'),
    },
  },
  {
    id: 'status',
    accessorKey: 'status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Status' />
    ),
    cell: ({ row }) =>
      row.original.is_running ? (
        <Badge
          variant='outline'
          className='border-teal-200 bg-teal-100/30 text-teal-900 dark:text-teal-200'
        >
          Running
        </Badge>
      ) : (
        <Badge
          variant='outline'
          className='bg-destructive/10 dark:bg-destructive/50 text-destructive dark:text-primary border-destructive/10'
        >
          Not Running
        </Badge>
      ),
    filterFn: (row, columnId, filterValue: string[]) => {
      const cellValue = row.getValue(columnId) as string[]
      return filterValue.some((val) => cellValue.includes(val))
    },
    meta: {
      className: cn('border-l border-r'),
    },
    enableHiding: false,
    enableSorting: false,
  },
  {
    id: 'actions',
    cell: InterfacesTableRowActions,
  },
]
