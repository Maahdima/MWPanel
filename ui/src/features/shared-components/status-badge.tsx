import { Badge } from '@/components/ui/badge.tsx'

interface Props {
  color: 'green' | 'yellow' | 'red' | 'gray'
  text: string
}

export function ColoredBadge({ color, text }: Props) {
  const colors = {
    green:
      'bg-green-600/10 text-green-600 dark:bg-green-400/10 dark:text-green-400',
    yellow:
      'bg-yellow-600/10 text-yellow-600 dark:bg-yellow-400/10 dark:text-yellow-400',
    red: 'bg-red-600/10 text-red-600 dark:bg-red-400/10 dark:text-red-400',
    gray: 'bg-slate-100/20 text-slate-200 dark:bg-slate-100/20 dark:text-slate-300',
  }

  return (
    <Badge
      className={`h-7 rounded-sm border-none capitalize ${colors[color]} focus-visible:ring-${color}-600/20 focus-visible:outline-none [a&]:hover:bg-${color}-600/5 dark:[a&]:hover:bg-${color}-400/5`}
    >
      <span
        className={`size-2 rounded-sm bg-${color}-400 dark:bg-${color}-400`}
        aria-hidden='true'
      />
      {text}
    </Badge>
  )
}
