import { CheckCircle2, XCircle, Clock, Loader2, FileCode } from 'lucide-react';
import { Badge } from '@/ui/badge';
import { Card } from '@/ui/card';
import type { CompilationResult } from '@/types/api';

interface StatusBarProps {
  isCompiling: boolean;
  result: CompilationResult | null;
  statusMessage: string;
}

export function StatusBar({ isCompiling, result, statusMessage }: StatusBarProps) {
  // Determine the current state
  const getStatus = () => {
    if (isCompiling) return 'compiling';
    if (result?.compiled) return 'success';
    if (result && !result.compiled) return 'failed';
    return 'idle';
  };

  const status = getStatus();

  // Status configurations
  type StatusConfig = {
    icon: typeof FileCode;
    label: string;
    color: string;
    badgeVariant: 'outline' | 'default' | 'destructive';
    borderColor: string;
    animate?: boolean;
  };

  const statusConfig: Record<string, StatusConfig> = {
    idle: {
      icon: FileCode,
      label: 'Ready',
      color: 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300',
      badgeVariant: 'outline' as const,
      borderColor: 'border-slate-300 dark:border-slate-700',
    },
    compiling: {
      icon: Loader2,
      label: 'Compiling',
      color: 'bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300',
      badgeVariant: 'default' as const,
      borderColor: 'border-blue-400 dark:border-blue-600',
      animate: true,
    },
    success: {
      icon: CheckCircle2,
      label: 'Success',
      color: 'bg-green-50 dark:bg-green-950 text-green-700 dark:text-green-300',
      badgeVariant: 'default' as const,
      borderColor: 'border-green-400 dark:border-green-600',
    },
    failed: {
      icon: XCircle,
      label: 'Failed',
      color: 'bg-red-50 dark:bg-red-950 text-red-700 dark:text-red-300',
      badgeVariant: 'destructive' as const,
      borderColor: 'border-red-400 dark:border-red-600',
    },
  };

  const config = statusConfig[status];
  const StatusIcon = config.icon;

  // Format duration from nanoseconds to human readable
  const formatDuration = (ns: number): string => {
    const ms = ns / 1_000_000;
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  return (
    <Card
      className={`sticky top-0 z-50 rounded-none border-x-0 border-t-0 border-b-2 ${config.borderColor} ${config.color} transition-all duration-300 shadow-md`}
    >
      <div className="container mx-auto max-w-7xl px-4 py-3">
        <div className="flex items-center justify-between gap-4">
          {/* Left section: Status indicator */}
          <div className="flex items-center gap-3">
            <StatusIcon
              className={`h-6 w-6 ${config.animate ? 'animate-spin' : ''}`}
            />
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <Badge variant={config.badgeVariant} className="text-sm font-semibold">
                  {config.label}
                </Badge>
                {statusMessage && (
                  <span className="text-sm font-medium">{statusMessage}</span>
                )}
              </div>
            </div>
          </div>

          {/* Right section: Job details */}
          {result && (
            <div className="flex items-center gap-3 flex-wrap">
              {/* Exit Code */}
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium opacity-70">Exit Code:</span>
                <Badge
                  variant={result.exit_code === 0 ? 'default' : 'destructive'}
                  className="font-mono"
                >
                  {result.exit_code}
                </Badge>
              </div>

              {/* Duration */}
              <div className="flex items-center gap-1.5">
                <Clock className="h-4 w-4 opacity-70" />
                <span className="text-sm font-medium">
                  {formatDuration(result.duration)}
                </span>
              </div>

              {/* Job ID */}
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium opacity-70">Job:</span>
                <code className="text-xs font-mono bg-black/10 dark:bg-white/10 px-2 py-1 rounded">
                  {result.job_id.substring(0, 8)}
                </code>
              </div>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}
