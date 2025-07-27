import React, { useState } from 'react'
import { Server } from '@/schema/servers.ts'
import useDialogState from '@/hooks/use-dialog-state'

type PoolsDialogType = 'add' | 'edit' | 'delete'

interface PoolsContextType {
  open: PoolsDialogType | null
  setOpen: (str: PoolsDialogType | null) => void
  currentRow: Server | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Server | null>>
}

const PoolsContext = React.createContext<PoolsContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function PoolsProvider({ children }: Props) {
  const [open, setOpen] = useDialogState<PoolsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Server | null>(null)

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
