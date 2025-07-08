import React, { useState } from 'react'
import { Server } from '@/schema/servers.ts'
import useDialogState from '@/hooks/use-dialog-state'

type ServersDialogType = 'add' | 'edit' | 'delete'

interface ServersContextType {
  open: ServersDialogType | null
  setOpen: (str: ServersDialogType | null) => void
  currentRow: Server | null
  setCurrentRow: React.Dispatch<React.SetStateAction<Server | null>>
}

const ServersContext = React.createContext<ServersContextType | null>(null)

interface Props {
  children: React.ReactNode
}

export default function ServersProvider({ children }: Props) {
  const [open, setOpen] = useDialogState<ServersDialogType>(null)
  const [currentRow, setCurrentRow] = useState<Server | null>(null)

  return (
    <ServersContext value={{ open, setOpen, currentRow, setCurrentRow }}>
      {children}
    </ServersContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useServers = () => {
  const serversContext = React.useContext(ServersContext)

  if (!serversContext) {
    throw new Error('useServers has to be used within <ServersContext>')
  }

  return serversContext
}
