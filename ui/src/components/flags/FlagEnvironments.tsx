import {
  AlertCircleIcon,
  CheckCircle2Icon,
  CircleAlertIcon
} from 'lucide-react';
import { useMemo } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentEnvironment,
  selectEnvironments
} from '~/app/environments/environmentsApi';
import { useGetFlagQuery } from '~/app/flags/flagsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { Badge } from '~/components/Badge';

import { IFlag } from '~/types/Flag';
import { IEnvironment } from '~/types/Environment';

import { driftFields } from '~/utils/compare';

function EnvironmentRow({
  environment,
  flagKey,
  namespaceKey,
  currentFlag
}: {
  environment: IEnvironment;
  flagKey: string;
  namespaceKey: string;
  currentFlag: IFlag;
}) {
  const { data: flag, isLoading, error } = useGetFlagQuery(
    {
      environmentKey: environment.key,
      namespaceKey,
      flagKey
    },
    { refetchOnFocus: false }
  );

  const drift = useMemo(() => {
    if (!flag) return [];
    return driftFields(
      currentFlag as unknown as Record<string, unknown>,
      flag as unknown as Record<string, unknown>
    );
  }, [flag, currentFlag]);

  if (isLoading) {
    return (
      <div className="flex flex-wrap items-center gap-4 p-3">
        <div className="w-32 font-medium text-sm">{environment.name || environment.key}</div>
        <div className="h-5 w-16 animate-pulse rounded bg-muted" />
        <div className="h-4 w-20 animate-pulse rounded bg-muted" />
        <div className="h-4 w-14 animate-pulse rounded bg-muted" />
      </div>
    );
  }

  if (error) {
    const isNotFound =
      error && typeof error === 'object' && 'status' in error && error.status === 404;

    return (
      <div className="flex items-center gap-4 p-3">
        <div className="w-32 font-medium text-sm">{environment.name || environment.key}</div>
        {isNotFound ? (
          <Badge variant="secondary" state="muted">
            Not Present
          </Badge>
        ) : (
          <div className="flex items-center gap-2 text-sm text-red-600">
            <AlertCircleIcon className="h-4 w-4" />
            Failed to load
          </div>
        )}
      </div>
    );
  }

  if (!flag) return null;

  const hasDrift = drift.length > 0;
  const enabledMatch = flag.enabled === currentFlag.enabled;

  return (
    <div className="flex flex-wrap items-center gap-4 p-3">
      <div className="w-32 font-medium text-sm truncate" title={environment.key}>
        {environment.name || environment.key}
      </div>

      <Badge
        variant="secondary"
        state={flag.enabled ? 'success' : 'muted'}
      >
        {flag.enabled ? 'Enabled' : 'Disabled'}
      </Badge>

      {!enabledMatch && (
        <span className="text-xs text-amber-600">
          (differs from current)
        </span>
      )}

      <span className="text-xs text-muted-foreground">
        {flag.variants?.length || 0} variants
      </span>

      <span className="text-xs text-muted-foreground">
        {flag.rules?.length || 0} rules
      </span>

      <span className="text-xs text-muted-foreground">
        {flag.rollouts?.length || 0} rollouts
      </span>

      <div className="ml-auto flex items-center gap-1">
        {hasDrift ? (
          <>
            <CircleAlertIcon className="h-4 w-4 text-amber-600" />
            <span className="text-xs text-amber-600" title={`Drift: ${drift.join(', ')}`}>
              Drift: {drift.join(', ')}
            </span>
          </>
        ) : (
          <>
            <CheckCircle2Icon className="h-4 w-4 text-emerald-600" />
            <span className="text-xs text-emerald-600">In sync</span>
          </>
        )}
      </div>
    </div>
  );
}

export default function FlagEnvironments({ flag }: { flag: IFlag }) {
  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const environments = useSelector(selectEnvironments);

  const otherEnvironments = useMemo(
    () => environments.filter((env) => env.key !== currentEnvironment.key),
    [environments, currentEnvironment.key]
  );

  if (otherEnvironments.length === 0) {
    return (
      <div className="mt-4 rounded-md border p-6 text-center text-sm text-muted-foreground">
        Only one environment configured. Add more environments to compare flag state across them.
      </div>
    );
  }

  return (
    <div className="mt-4 rounded-md border">
      <div className="border-b p-3">
        <div className="text-sm font-medium">
          Flag state across environments
        </div>
        <div className="text-xs text-muted-foreground">
          Comparing against current environment: {currentEnvironment.name || currentEnvironment.key}
        </div>
      </div>

      {/* Current environment row */}
      <div className="flex flex-wrap items-center gap-4 bg-muted/30 p-3">
        <div className="w-32 font-medium text-sm truncate" title={currentEnvironment.key}>
          {currentEnvironment.name || currentEnvironment.key}
        </div>

        <Badge
          variant="secondary"
          state={flag.enabled ? 'success' : 'muted'}
        >
          {flag.enabled ? 'Enabled' : 'Disabled'}
        </Badge>

        <span className="text-xs text-muted-foreground">
          {flag.variants?.length || 0} variants
        </span>

        <span className="text-xs text-muted-foreground">
          {flag.rules?.length || 0} rules
        </span>

        <span className="text-xs text-muted-foreground">
          {flag.rollouts?.length || 0} rollouts
        </span>

        <div className="ml-auto">
          <Badge variant="outlinemuted">Current</Badge>
        </div>
      </div>

      {/* Other environment rows */}
      <div className="divide-y">
        {otherEnvironments.map((env) => (
          <EnvironmentRow
            key={env.key}
            environment={env}
            flagKey={flag.key}
            namespaceKey={namespace.key}
            currentFlag={flag}
          />
        ))}
      </div>
    </div>
  );
}
