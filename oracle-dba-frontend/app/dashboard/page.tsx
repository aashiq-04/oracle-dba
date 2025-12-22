'use client';

import { useEffect } from 'react';
import { useQuery } from '@apollo/client/react';
import Navbar from '../components/Navbar';
import { requireAuth } from '@/lib/auth';
import {
  SESSION_SUMMARY_QUERY,
  TABLESPACES_QUERY,
  BLOCKING_SESSIONS_QUERY,
  DATABASE_INSTANCE_QUERY,
} from '@/lib/queries';

import {
  SessionSummary,
  Tablespace,
  BlockingSession,
  DatabaseInstance,
  SessionsBySchema,
} from '@/types';

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';

/* ---------------- Query Result Types ---------------- */

interface SessionSummaryQuery {
  sessionSummary: SessionSummary;
}

interface TablespacesQuery {
  tablespaces: Tablespace[];
}

interface BlockingSessionsQuery {
  blockingSessions: BlockingSession[];
}

interface DatabaseInstanceQuery {
  databaseInstance: DatabaseInstance;
}

/* ---------------- Component ---------------- */

export default function DashboardPage() {
  useEffect(() => {
    requireAuth();
  }, []);

  const { data: sessionData, loading: sessionLoading } =
    useQuery<SessionSummaryQuery>(SESSION_SUMMARY_QUERY, {
      pollInterval: 10_000,
    });

  const { data: tablespaceData, loading: tablespaceLoading } =
    useQuery<TablespacesQuery>(TABLESPACES_QUERY, {
      pollInterval: 30_000,
    });

  const { data: blockingData } =
    useQuery<BlockingSessionsQuery>(BLOCKING_SESSIONS_QUERY, {
      pollInterval: 5_000,
    });

  const { data: dbInstanceData } =
    useQuery<DatabaseInstanceQuery>(DATABASE_INSTANCE_QUERY);

  /* ---------------- Derived Data ---------------- */

  const tablespaceChartData =
    tablespaceData?.tablespaces.map((ts: Tablespace) => ({
      name: ts?.name,
      usage: ts?.usagePercentage,
    })) ?? [];

  const getBarColor = (usage: number): string => {
    if (usage >= 90) return '#ef4444';
    if (usage >= 75) return '#f59e0b';
    return '#10b981';
  };

  /* ---------------- Render ---------------- */

  return (
    <div className="min-h-screen bg-gray-50">
      <Navbar />

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-600 mt-1">
            Real-time Oracle database monitoring
          </p>
        </div>

        {/* Database Instance */}
        {dbInstanceData?.databaseInstance && (
          <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-lg font-semibold mb-4">Database Instance</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <Info label="Instance" value={dbInstanceData.databaseInstance.instanceName} />
              <Info label="Version" value={dbInstanceData.databaseInstance.version} />
              <Info label="Status" value={dbInstanceData.databaseInstance.status} />
              <Info
                label="Uptime"
                value={`${dbInstanceData.databaseInstance.uptimeDays.toFixed(1)} days`}
              />
            </div>
          </div>
        )}

        {/* Session Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <StatCard
            title="Total Sessions"
            value={sessionData?.sessionSummary.totalSessions ?? 0}
            icon="👥"
            color="blue"
            loading={sessionLoading}
          />
          <StatCard
            title="Active Sessions"
            value={sessionData?.sessionSummary.activeSessions ?? 0}
            icon="⚡"
            color="green"
            loading={sessionLoading}
          />
          <StatCard
            title="Inactive Sessions"
            value={sessionData?.sessionSummary.inactiveSessions ?? 0}
            icon="💤"
            color="gray"
            loading={sessionLoading}
          />
          <StatCard
            title="Blocked Sessions"
            value={sessionData?.sessionSummary.blockedSessions ?? 0}
            icon="🚫"
            color="red"
            loading={sessionLoading}
            alert={(sessionData?.sessionSummary.blockedSessions ?? 0) > 0}
          />
        </div>

        {/* Blocking Alert */}
        {blockingData?.blockingSessions.length ? (
          <div className="bg-red-50 border-l-4 border-red-500 p-4 mb-6">
            <p className="font-medium text-red-800">
              {blockingData.blockingSessions.length} blocking session(s) detected
            </p>
          </div>
        ) : null}

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Tablespace Chart */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold mb-4">Tablespace Usage</h2>

            {tablespaceLoading ? (
              <Loader />
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={tablespaceChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" angle={-45} textAnchor="end" height={100} />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="usage">
                    {tablespaceChartData.map((entry: { name: string; usage: number }, index: number) => (
                      <Cell
                        key={index}
                        fill={getBarColor(entry.usage)}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>

          {/* Sessions by Schema */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold mb-4">Sessions by Schema</h2>

            {sessionLoading ? (
              <Loader />
            ) : (
              <div className="space-y-3">
                {sessionData?.sessionSummary.bySchema
                  .slice(0, 8)
                  .map((schema: SessionsBySchema) => (
                    <div
                      key={schema?.schemaName}
                      className="flex justify-between items-center"
                    >
                      <div>
                        <p className="font-medium">{schema?.schemaName}</p>
                        <p className="text-xs text-gray-600">
                          Total {schema?.total} · Active {schema?.active} · Inactive{' '}
                          {schema?.inactive}
                        </p>
                      </div>
                      <span className="text-lg font-bold">{schema?.total}</span>
                    </div>
                  ))}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

/* ---------------- Small Components ---------------- */

function Loader() {
  return (
    <div className="h-64 flex items-center justify-center">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-lg font-medium">{value}</p>
    </div>
  );
}

function StatCard({
  title,
  value,
  icon,
  color,
  loading,
  alert,
}: {
  title: string;
  value: number;
  icon: string;
  color: 'blue' | 'green' | 'red' | 'gray';
  loading?: boolean;
  alert?: boolean;
}) {
  const colorMap = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    red: 'bg-red-50 text-red-600',
    gray: 'bg-gray-50 text-gray-600',
  };

  return (
    <div className={`bg-white rounded-lg shadow p-6 ${alert ? 'ring-2 ring-red-500' : ''}`}>
      <div className="flex justify-between items-center">
        <div>
          <p className="text-sm text-gray-600">{title}</p>
          {loading ? (
            <div className="h-8 w-16 bg-gray-200 rounded animate-pulse" />
          ) : (
            <p className="text-3xl font-bold">{value}</p>
          )}
        </div>
        <div className={`w-12 h-12 rounded-full flex items-center justify-center ${colorMap[color]}`}>
          {icon}
        </div>
      </div>
    </div>
  );
}
