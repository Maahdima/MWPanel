import React, { useState } from 'react'
import { IPPool } from '@/schema/ip-pool.ts'
import useDialogState from '@/hooks/use-dialog-state'

type PoolsDialogType = 'add' | 'edit' | 'delete'

interface PoolsContextType {
  open: PoolsDialogType | null
  setOpen: (str: PoolsDialogType | null) => void
  currentRow: IPPool | null
  setCurrentRow: React.Dispatch<React.SetStateAction<IPPool | null>>
}

const PoolsContext = React.createContext<PoolsContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function PoolsProvider({ children }: Props) {
  const [open, setOpen] = useDialogState<PoolsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<IPPool | null>(null)

  return (
    <PoolsContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </PoolsContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const usePools = () => {
  const poolsContext = React.useContext(PoolsContext)

  if (!poolsContext) {
    throw new Error('usePools has to be used within <PoolsContext>')
  }

  return poolsContext
}
