import { useId, useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

function SpeedField({
  name,
  label,
  initial,
}: {
  name: string
  label: string
  initial?: string | null
}) {
  const id = useId()
  const match = initial?.match(/^(\d+(?:\.\d+)?)([KMG])$/)
  const [amount, setAmount] = useState(match?.[1] ?? '')
  const [unit, setUnit] = useState(match?.[2] ?? 'M')
  return (
    <div className='space-y-2'>
      <Label htmlFor={id}>{label}</Label>
      <input
        type='hidden'
        name={name}
        value={amount ? `${amount}${unit}` : ''}
      />
      <div className='flex min-w-0 gap-2'>
        <Input
          id={id}
          type='number'
          min='0.001'
          step='0.001'
          placeholder='Unlimited'
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
        />
        <Select value={unit} onValueChange={setUnit}>
          <SelectTrigger className='w-28 shrink-0' aria-label={`${label} unit`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value='K'>Kbit/s</SelectItem>
            <SelectItem value='M'>Mbit/s</SelectItem>
            <SelectItem value='G'>Gbit/s</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}

export function BandwidthFields({
  download,
  upload,
}: {
  download?: string | null
  upload?: string | null
}) {
  return (
    <div className='space-y-2'>
      <div className='grid gap-4 sm:grid-cols-2'>
        <SpeedField
          name='download_bandwidth'
          label='Download speed'
          initial={download}
        />
        <SpeedField
          name='upload_bandwidth'
          label='Upload speed'
          initial={upload}
        />
      </div>
      <p className='text-muted-foreground text-sm'>
        Leave blank for unlimited speed. Speed limits apply to this
        configuration only and do not change its price.
      </p>
    </div>
  )
}
