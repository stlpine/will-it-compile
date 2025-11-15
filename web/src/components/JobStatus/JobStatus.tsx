import { Alert, AlertDescription, AlertTitle } from '@/ui/alert'
import { Loader2 } from 'lucide-react'

interface JobStatusProps {
  isCompiling: boolean
  statusMessage: string
}

/**
 * Component to display job status during compilation with shadcn/ui
 */
export function JobStatus({ isCompiling, statusMessage }: JobStatusProps) {
  if (!isCompiling && !statusMessage) {
    return null
  }

  return (
    <Alert variant="info" className="animate-in fade-in slide-in-from-top-2">
      <Loader2 className={`h-5 w-5 ${isCompiling ? 'animate-spin' : ''}`} />
      <AlertTitle>{isCompiling ? 'Compiling...' : 'Status'}</AlertTitle>
      {statusMessage && <AlertDescription>{statusMessage}</AlertDescription>}
    </Alert>
  )
}
