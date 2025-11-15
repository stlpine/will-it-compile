import { CompilationResult } from '@/types/api.ts'
import { formatDuration } from '@/utils/formatters.ts'
import { sanitizeOutput } from '@/utils/security.ts'
import { Alert, AlertDescription, AlertTitle } from '@/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card'
import { Badge } from '@/ui/badge'
import {
  CheckCircle2,
  XCircle,
  AlertTriangle,
  FileText,
  Terminal,
} from 'lucide-react'
import { Separator } from '@/ui/separator'

interface CompilerOutputProps {
  result: CompilationResult
}

/**
 * Component to display compilation results with shadcn/ui
 * Sanitizes all output to prevent XSS attacks
 */
export function CompilerOutput({ result }: CompilerOutputProps) {
  const hasStdout = result.stdout && result.stdout.trim().length > 0
  const hasStderr = result.stderr && result.stderr.trim().length > 0

  // Sanitize outputs to remove dangerous control sequences
  const safeStdout = sanitizeOutput(result.stdout || '')
  const safeStderr = sanitizeOutput(result.stderr || '')
  const safeError = sanitizeOutput(result.error || '')

  return (
    <div className="space-y-4">
      {/* Status Banner */}
      <Alert variant={result.compiled ? 'success' : 'destructive'}>
        {result.compiled ? (
          <CheckCircle2 className="h-5 w-5" />
        ) : (
          <XCircle className="h-5 w-5" />
        )}
        <AlertTitle className="text-lg">
          {result.compiled ? 'Compilation Successful ✨' : 'Compilation Failed'}
        </AlertTitle>
        <AlertDescription className="mt-2 flex flex-col gap-2">
          <div className="flex items-center gap-4 flex-wrap">
            <Badge variant={result.exit_code === 0 ? 'success' : 'destructive'}>
              Exit Code: {result.exit_code}
            </Badge>
            <Badge variant="outline">
              Duration: {formatDuration(result.duration)}
            </Badge>
            <Badge variant="outline" className="font-mono text-xs">
              Job: {result.job_id.substring(0, 8)}...
            </Badge>
          </div>
        </AlertDescription>
      </Alert>

      {/* Error Message */}
      {result.error && (
        <Alert variant="destructive">
          <AlertTriangle className="h-5 w-5" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>
            <pre className="mt-2 text-sm whitespace-pre-wrap font-mono">
              {safeError}
            </pre>
          </AlertDescription>
        </Alert>
      )}

      {/* Standard Output */}
      {hasStdout && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Terminal className="h-4 w-4" />
              Standard Output
            </CardTitle>
          </CardHeader>
          <Separator />
          <CardContent className="pt-4">
            <pre className="text-sm whitespace-pre-wrap font-mono bg-muted p-4 rounded-md max-h-96 overflow-y-auto scrollbar-thin">
              {safeStdout}
            </pre>
          </CardContent>
        </Card>
      )}

      {/* Standard Error */}
      {hasStderr && (
        <Card
          className={
            result.compiled
              ? 'border-yellow-200 bg-yellow-50/50'
              : 'border-destructive/50'
          }
        >
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              {result.compiled ? 'Warnings' : 'Compilation Errors'}
            </CardTitle>
          </CardHeader>
          <Separator />
          <CardContent className="pt-4">
            <pre
              className={`text-sm whitespace-pre-wrap font-mono p-4 rounded-md max-h-96 overflow-y-auto scrollbar-thin ${
                result.compiled
                  ? 'bg-yellow-100/50 text-yellow-900'
                  : 'bg-destructive/5 text-destructive'
              }`}
            >
              {safeStderr}
            </pre>
          </CardContent>
        </Card>
      )}

      {/* No Output Message */}
      {!hasStdout && !hasStderr && !result.error && (
        <Card>
          <CardContent className="pt-6 pb-6 text-center">
            <FileText className="h-8 w-8 mx-auto text-muted-foreground mb-2" />
            <p className="text-sm text-muted-foreground">No output generated</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
