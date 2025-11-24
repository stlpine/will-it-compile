import { useEffect, useState, forwardRef, useImperativeHandle } from 'react'
import { Activity, CheckCircle, Clock, XCircle, Users, Server, AlertTriangle, AlertCircle } from 'lucide-react'
import { getWorkerStats } from '../../services/api'
import type { WorkerStats } from '../../types/api'

interface WorkerPoolStatusProps {
  autoRefresh?: boolean
  refreshInterval?: number // in milliseconds
}

export interface WorkerPoolStatusHandle {
  refresh: () => Promise<void>
}

export const WorkerPoolStatus = forwardRef<WorkerPoolStatusHandle, WorkerPoolStatusProps>(
  function WorkerPoolStatus({ autoRefresh = true, refreshInterval = 2000 }, ref) {
  const [stats, setStats] = useState<WorkerStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchStats = async () => {
    try {
      const data = await getWorkerStats()
      setStats(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch worker stats')
    } finally {
      setLoading(false)
    }
  }

  // Expose refresh method to parent components
  useImperativeHandle(ref, () => ({
    refresh: fetchStats,
  }))

  useEffect(() => {
    fetchStats()

    if (autoRefresh) {
      const interval = setInterval(fetchStats, refreshInterval)
      return () => clearInterval(interval)
    }
  }, [autoRefresh, refreshInterval])

  if (loading) {
    return (
      <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-200 border-t-blue-600" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-4">
        <div className="flex items-center gap-2 text-red-700">
          <XCircle className="h-5 w-5" />
          <span className="font-medium">Error loading worker stats</span>
        </div>
        <p className="mt-1 text-sm text-red-600">{error}</p>
      </div>
    )
  }

  if (!stats) {
    return null
  }

  const utilizationPercent = (stats.active_workers / stats.max_workers) * 100
  const successRate =
    stats.total_processed > 0
      ? (stats.total_successful / stats.total_processed) * 100
      : 0

  return (
    <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
      {/* Header */}
      <div className="border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Server className="h-5 w-5 text-blue-600" />
            <h2 className="text-lg font-semibold text-gray-900">Worker Pool Status</h2>
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <Clock className="h-4 w-4" />
            <span>Uptime: {stats.uptime}</span>
          </div>
        </div>
      </div>

      {/* Main Stats Grid */}
      <div className="grid grid-cols-1 gap-4 p-6 sm:grid-cols-2 lg:grid-cols-4">
        {/* Max Workers */}
        <StatCard
          icon={<Users className="h-5 w-5 text-blue-600" />}
          label="Total Workers"
          value={stats.max_workers}
          description="Maximum capacity"
        />

        {/* Active Workers */}
        <StatCard
          icon={<Activity className="h-5 w-5 text-green-600" />}
          label="Active Workers"
          value={stats.active_workers}
          description={`${utilizationPercent.toFixed(0)}% utilization`}
          highlight={stats.active_workers > 0}
        />

        {/* Available Slots */}
        <StatCard
          icon={<CheckCircle className="h-5 w-5 text-emerald-600" />}
          label="Available Slots"
          value={stats.available_slots}
          description="Ready for jobs"
          highlight={stats.available_slots > 0}
        />

        {/* Queued Jobs */}
        <StatCard
          icon={<Clock className="h-5 w-5 text-amber-600" />}
          label="Queued Jobs"
          value={stats.queued_jobs}
          description="Waiting to process"
          highlight={stats.queued_jobs > 0}
          warning={stats.queued_jobs > stats.max_workers}
        />
      </div>

      {/* Utilization Bar */}
      <div className="px-6 pb-4">
        <div className="rounded-lg bg-gray-100 p-4">
          <div className="mb-2 flex items-center justify-between text-sm">
            <span className="font-medium text-gray-700">Pool Utilization</span>
            <span className="font-semibold text-gray-900">
              {stats.active_workers} / {stats.max_workers}
            </span>
          </div>
          <div className="h-3 w-full overflow-hidden rounded-full bg-gray-200">
            <div
              className={`h-full transition-all duration-500 ${
                utilizationPercent >= 90
                  ? 'bg-red-600'
                  : utilizationPercent >= 70
                    ? 'bg-amber-500'
                    : 'bg-green-500'
              }`}
              style={{ width: `${utilizationPercent}%` }}
            />
          </div>
        </div>
      </div>

      {/* Processing Stats */}
      <div className="border-t border-gray-200 px-6 py-4">
        <h3 className="mb-3 text-sm font-semibold text-gray-700">
          Processing Statistics
        </h3>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-5">
          <div className="flex items-center justify-between rounded-md bg-gray-50 px-3 py-2">
            <span className="text-sm text-gray-600">Total</span>
            <span className="font-semibold text-gray-900">{stats.total_processed}</span>
          </div>
          <div className="flex items-center justify-between rounded-md bg-green-50 px-3 py-2">
            <div className="flex items-center gap-1">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <span className="text-sm text-green-700">Compiled</span>
            </div>
            <span className="font-semibold text-green-900">
              {stats.total_successful}
            </span>
          </div>
          <div className="flex items-center justify-between rounded-md bg-red-50 px-3 py-2">
            <div className="flex items-center gap-1">
              <XCircle className="h-4 w-4 text-red-600" />
              <span className="text-sm text-red-700">Failed</span>
            </div>
            <span className="font-semibold text-red-900">{stats.total_failed}</span>
          </div>
          <div className="flex items-center justify-between rounded-md bg-amber-50 px-3 py-2">
            <div className="flex items-center gap-1">
              <Clock className="h-4 w-4 text-amber-600" />
              <span className="text-sm text-amber-700">Timeout</span>
            </div>
            <span className="font-semibold text-amber-900">{stats.total_timeout}</span>
          </div>
          <div className="flex items-center justify-between rounded-md bg-orange-50 px-3 py-2">
            <div className="flex items-center gap-1">
              <AlertCircle className="h-4 w-4 text-orange-600" />
              <span className="text-sm text-orange-700">Errors</span>
            </div>
            <span className="font-semibold text-orange-900">{stats.total_errors}</span>
          </div>
        </div>
        {stats.total_processed > 0 && (
          <div className="mt-3 text-center text-sm text-gray-600">
            Compilation Success Rate: <span className="font-semibold">{successRate.toFixed(1)}%</span>
            {stats.total_errors > 0 && (
              <span className="ml-4 text-orange-600">
                System Reliability: <span className="font-semibold">
                  {((stats.total_processed - stats.total_errors) / stats.total_processed * 100).toFixed(1)}%
                </span>
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  )
})

interface StatCardProps {
  icon: React.ReactNode
  label: string
  value: number
  description: string
  highlight?: boolean
  warning?: boolean
}

function StatCard({ icon, label, value, description, highlight, warning }: StatCardProps) {
  return (
    <div
      className={`rounded-lg border p-4 transition-colors ${
        warning
          ? 'border-amber-200 bg-amber-50'
          : highlight
            ? 'border-blue-200 bg-blue-50'
            : 'border-gray-200 bg-gray-50'
      }`}
    >
      <div className="flex items-center gap-3">
        <div className="rounded-md bg-white p-2 shadow-sm">{icon}</div>
        <div className="flex-1">
          <p className="text-sm text-gray-600">{label}</p>
          <p className={`text-2xl font-bold ${warning ? 'text-amber-900' : 'text-gray-900'}`}>
            {value}
          </p>
        </div>
      </div>
      <p className={`mt-2 text-xs ${warning ? 'text-amber-700' : 'text-gray-500'}`}>
        {description}
      </p>
    </div>
  )
}
