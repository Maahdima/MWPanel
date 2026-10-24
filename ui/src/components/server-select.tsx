import { useServersListQuery } from '@/hooks/servers/useServersListQuery'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface Props {
  value?: number
  onValueChange: (value: number) => void
  disabled?: boolean
}

export function ServerSelect({ value, onValueChange, disabled }: Props) {
  const { data: servers = [], isLoading, isError } = useServersListQuery()
  const available = servers.filter((server) => server.is_active)
  const placeholder = isLoading
    ? 'Loading servers...'
    : isError
      ? 'Unable to load servers'
      : available.length
        ? 'Select server'
        : 'No active servers'

  return (
    <Select
      value={value?.toString() ?? ''}
      onValueChange={(selected) => onValueChange(Number(selected))}
      disabled={disabled || isLoading || isError || !available.length}
    >
      <SelectTrigger aria-label='Server' className='w-full'>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {available.map((server) => (
          <SelectItem key={server.id} value={server.id.toString()}>
            {server.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

export function ServerName({ id }: { id: number }) {
  const { data: servers = [] } = useServersListQuery()
  return (
    <span>
      {servers.find((server) => server.id === id)?.name ??
        (id ? `Server ${id}` : 'Unassigned')}
    </span>
  )
}
