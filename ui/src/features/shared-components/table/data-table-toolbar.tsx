import { Cross2Icon } from '@radix-ui/react-icons'
import { Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button.tsx'
import { Input } from '@/components/ui/input.tsx'
import { DataTableFacetedFilter } from './data-table-faceted-filter.tsx'
import { DataTableViewOptions } from './data-table-view-options.tsx'

interface DataTableToolbarProps<TData> {
  table: Table<TData>
  type: 'Servers' | 'Interfaces' | 'Pools' | 'Peers'
}

const serverStatusOptions = [
  { label: 'Available', value: 'available' },
  { label: 'Not Available', value: 'not_available' },
]

const interfaceStatusOptions = [
  { label: 'Running', value: 'running' },
  { label: 'Not Running', value: 'not_running' },
]

const peerStatusOptions = [
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
  { label: 'Expired', value: 'expired' },
  { label: 'Suspended', value: 'suspended' },
]

export function DataTableToolbar<TData>({
  table,
  type,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0

  return (
    <div className='flex items-center justify-between'>
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:items-center sm:space-x-2'>
        <Input
          placeholder={`Filter ${type}...`}
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(event) =>
            table.getColumn('name')?.setFilterValue(event.target.value)
          }
          className='h-8 w-[150px] lg:w-[250px]'
        />
        <div className='flex gap-x-2'>
          {type !== 'Pools' && table.getColumn('status') && (
            <DataTableFacetedFilter
              column={table.getColumn('status')}
              title='Status'
              options={
                type === 'Servers'
                  ? serverStatusOptions
                  : type === 'Interfaces'
                    ? interfaceStatusOptions
                    : peerStatusOptions
              }
            />
          )}
        </div>
        {isFiltered && (
          <Button
            variant='ghost'
            onClick={() => table.resetColumnFilters()}
            className='h-8 px-2 lg:px-3'
          >
            Reset
            <Cross2Icon className='ml-2 h-4 w-4' />
          </Button>
        )}
      </div>
      <DataTableViewOptions table={table} />
    </div>
  )
}
