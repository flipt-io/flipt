import {
  CheckCircle2Icon,
  CircleAlertIcon,
  CopyIcon,
  MinusCircleIcon
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectAllEnvironments,
  selectCurrentEnvironment,
  useBulkApplyResourcesMutation,
  useCompareEnvironmentsQuery
} from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';
import {
  selectCurrentNamespace,
  useListNamespacesQuery
} from '~/app/namespaces/namespacesApi';

import { Badge } from '~/components/Badge';
import { Button } from '~/components/Button';
import { PageHeader } from '~/components/Page';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/Select';

import { Product } from '~/types/Meta';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { useNotification } from '~/data/hooks/notification';
import {
  type CompareStatus,
  diffFlag,
  diffSegment,
  toResourceType,
  toStatus
} from '~/utils/compare';
import { getRevision } from '~/utils/helpers';

type CompareFilter = 'all' | 'different' | 'missing';

type CompareRow = {
  id: string;
  resourceType: 'flag' | 'segment';
  key: string;
  status: CompareStatus;
  differences: string[];
  sourcePayload: Record<string, unknown> | undefined;
};

const TYPE_URL_BY_RESOURCE = {
  flag: 'flipt.core.Flag',
  segment: 'flipt.core.Segment'
} as const;

function statusBadge(status: CompareStatus) {
  if (status === 'same') {
    return (
      <Badge variant="secondary" state="muted">
        Same
      </Badge>
    );
  }

  if (status === 'different') {
    return (
      <Badge variant="secondary" state="error">
        Different
      </Badge>
    );
  }

  if (status === 'source_only') {
    return (
      <Badge variant="secondary" state="success">
        Missing In Target
      </Badge>
    );
  }

  return (
    <Badge variant="secondary" state="muted">
      Source Missing
    </Badge>
  );
}

function defaultNamespaceKey(
  namespaces: INamespace[] | undefined,
  fallback: string
) {
  if (!namespaces || namespaces.length === 0) {
    return fallback;
  }

  const matched = namespaces.find((ns) => ns.key === fallback);
  return matched?.key || namespaces[0].key;
}

export default function Compare() {
  const info = useSelector(selectInfo);
  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const currentNamespace = useSelector(selectCurrentNamespace);
  const environments = useSelector(selectAllEnvironments);

  const [sourceEnvironmentKey, setSourceEnvironmentKey] = useState(
    currentEnvironment.key
  );
  const [targetEnvironmentKey, setTargetEnvironmentKey] = useState(() => {
    const other = environments.find(
      (env) => env.key !== currentEnvironment.key
    );
    return other?.key || currentEnvironment.key;
  });
  const [sourceNamespaceKey, setSourceNamespaceKey] = useState(
    currentNamespace.key
  );
  const [targetNamespaceKey, setTargetNamespaceKey] = useState(
    currentNamespace.key
  );
  const [filter, setFilter] = useState<CompareFilter>('all');
  const [selectedRows, setSelectedRows] = useState<Record<string, boolean>>({});
  const [conflictStrategy, setConflictStrategy] = useState('OVERWRITE');

  const sourceNamespacesQuery = useListNamespacesQuery({
    environmentKey: sourceEnvironmentKey
  });
  const targetNamespacesQuery = useListNamespacesQuery({
    environmentKey: targetEnvironmentKey
  });

  const sourceNamespaces = useMemo(
    () => sourceNamespacesQuery.data?.items || [],
    [sourceNamespacesQuery.data?.items]
  );
  const targetNamespaces = useMemo(
    () => targetNamespacesQuery.data?.items || [],
    [targetNamespacesQuery.data?.items]
  );

  useEffect(() => {
    setSourceNamespaceKey((current) =>
      defaultNamespaceKey(sourceNamespaces, current)
    );
  }, [sourceEnvironmentKey, sourceNamespaces]);

  useEffect(() => {
    setTargetNamespaceKey((current) =>
      defaultNamespaceKey(targetNamespaces, current)
    );
  }, [targetEnvironmentKey, targetNamespaces]);

  const compareQuery = useCompareEnvironmentsQuery(
    {
      environmentKey: sourceEnvironmentKey,
      namespaceKey: sourceNamespaceKey,
      targetEnvironmentKey,
      targetNamespaceKey
    },
    { skip: !sourceNamespaceKey || !targetNamespaceKey }
  );

  const { setNotification } = useNotification();
  const { setError } = useError();
  const [bulkApplyResources, { isLoading: isCopying }] =
    useBulkApplyResourcesMutation();

  const rows = useMemo<CompareRow[]>(() => {
    const items = compareQuery.data?.results || [];
    const mapped: CompareRow[] = [];

    for (const result of items) {
      const resourceType = toResourceType(result.typeUrl);
      if (!resourceType) {
        continue;
      }

      const status = toStatus(result.status);
      const sourcePayload = result.source?.payload as
        | Record<string, unknown>
        | undefined;
      const targetPayload = result.target?.payload as
        | Record<string, unknown>
        | undefined;

      let differences: string[] = [];
      if (status === 'different' && sourcePayload && targetPayload) {
        differences =
          resourceType === 'flag'
            ? diffFlag(sourcePayload, targetPayload)
            : diffSegment(sourcePayload, targetPayload);
      } else if (status === 'source_only') {
        differences = ['missing in target'];
      } else if (status === 'target_only') {
        differences = ['missing in source'];
      }

      mapped.push({
        id: `${resourceType}:${result.key}`,
        resourceType,
        key: result.key,
        status,
        differences,
        sourcePayload
      });
    }

    return mapped;
  }, [compareQuery.data?.results]);

  const visibleRows = useMemo(() => {
    if (filter === 'different') {
      return rows.filter((row) => row.status === 'different');
    }
    if (filter === 'missing') {
      return rows.filter((row) => row.status === 'source_only');
    }
    return rows;
  }, [filter, rows]);

  useEffect(() => {
    setSelectedRows((current) => {
      const next: Record<string, boolean> = {};
      for (const row of rows) {
        if (current[row.id] && row.status !== 'target_only') {
          next[row.id] = true;
        }
      }
      return next;
    });
  }, [rows]);

  const selected = useMemo(
    () => rows.filter((row) => selectedRows[row.id] && row.sourcePayload),
    [rows, selectedRows]
  );

  const statusCounts = useMemo(() => {
    return rows.reduce(
      (acc, row) => {
        acc[row.status] += 1;
        return acc;
      },
      {
        same: 0,
        different: 0,
        source_only: 0,
        target_only: 0
      } as Record<CompareStatus, number>
    );
  }, [rows]);

  const isLoading =
    sourceNamespacesQuery.isLoading ||
    targetNamespacesQuery.isLoading ||
    compareQuery.isLoading;

  const hasErrors =
    sourceNamespacesQuery.error ||
    targetNamespacesQuery.error ||
    compareQuery.error;

  useEffect(() => {
    if (hasErrors) {
      setError(hasErrors);
    }
  }, [hasErrors, setError]);

  const toggleSelect = (id: string, checked: boolean) => {
    setSelectedRows((current) => ({
      ...current,
      [id]: checked
    }));
  };

  const copySelected = async () => {
    if (selected.length === 0) {
      return;
    }

    if (info.product !== Product.PRO) {
      setNotification('Selective copy from compare view requires Flipt Pro.');
      return;
    }

    let revision =
      targetNamespacesQuery.data?.revision ||
      sourceNamespacesQuery.data?.revision ||
      getRevision();

    try {
      for (const row of selected) {
        if (!row.sourcePayload) {
          continue;
        }

        const response = await bulkApplyResources({
          environmentKey: targetEnvironmentKey,
          namespaceKeys: [targetNamespaceKey],
          operation: 'BULK_OPERATION_CREATE',
          typeUrl: TYPE_URL_BY_RESOURCE[row.resourceType],
          key: row.key,
          payload: {
            '@type': TYPE_URL_BY_RESOURCE[row.resourceType],
            ...row.sourcePayload
          },
          onConflict: conflictStrategy,
          revision
        }).unwrap();

        revision = response.revision;
      }

      setNotification(
        `Copied ${selected.length} resource${selected.length === 1 ? '' : 's'} to ${targetEnvironmentKey}/${targetNamespaceKey}.`
      );
      setSelectedRows({});
      compareQuery.refetch();
    } catch (error) {
      setError(error);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Compare Environments" />

      <div className="grid grid-cols-1 gap-4 rounded-md border p-4 md:grid-cols-2">
        <div className="space-y-2">
          <div className="text-sm font-medium">Source</div>
          <div className="grid grid-cols-1 gap-2">
            <Select
              value={sourceEnvironmentKey}
              onValueChange={setSourceEnvironmentKey}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select source environment" />
              </SelectTrigger>
              <SelectContent>
                {environments.map((environment) => (
                  <SelectItem key={environment.key} value={environment.key}>
                    {environment.key}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={sourceNamespaceKey}
              onValueChange={setSourceNamespaceKey}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select source namespace" />
              </SelectTrigger>
              <SelectContent>
                {sourceNamespaces.map((namespace) => (
                  <SelectItem key={namespace.key} value={namespace.key}>
                    {namespace.key}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="space-y-2">
          <div className="text-sm font-medium">Target</div>
          <div className="grid grid-cols-1 gap-2">
            <Select
              value={targetEnvironmentKey}
              onValueChange={setTargetEnvironmentKey}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select target environment" />
              </SelectTrigger>
              <SelectContent>
                {environments.map((environment) => (
                  <SelectItem key={environment.key} value={environment.key}>
                    {environment.key}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={targetNamespaceKey}
              onValueChange={setTargetNamespaceKey}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select target namespace" />
              </SelectTrigger>
              <SelectContent>
                {targetNamespaces.map((namespace) => (
                  <SelectItem key={namespace.key} value={namespace.key}>
                    {namespace.key}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Button
          variant={filter === 'all' ? 'primary' : 'outline'}
          onClick={() => setFilter('all')}
        >
          All ({rows.length})
        </Button>
        <Button
          variant={filter === 'different' ? 'primary' : 'outline'}
          onClick={() => setFilter('different')}
        >
          Different ({statusCounts.different})
        </Button>
        <Button
          variant={filter === 'missing' ? 'primary' : 'outline'}
          onClick={() => setFilter('missing')}
        >
          Missing In Target ({statusCounts.source_only})
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <div className="rounded-md border p-3">
          <div className="text-xs text-muted-foreground">Same</div>
          <div className="mt-1 text-xl font-semibold">{statusCounts.same}</div>
        </div>
        <div className="rounded-md border p-3">
          <div className="text-xs text-muted-foreground">Different</div>
          <div className="mt-1 text-xl font-semibold">
            {statusCounts.different}
          </div>
        </div>
        <div className="rounded-md border p-3">
          <div className="text-xs text-muted-foreground">Missing In Target</div>
          <div className="mt-1 text-xl font-semibold">
            {statusCounts.source_only}
          </div>
        </div>
        <div className="rounded-md border p-3">
          <div className="text-xs text-muted-foreground">Missing In Source</div>
          <div className="mt-1 text-xl font-semibold">
            {statusCounts.target_only}
          </div>
        </div>
      </div>

      <div className="rounded-md border">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b p-3">
          <div className="text-sm text-muted-foreground">
            {isLoading
              ? 'Loading comparison...'
              : `${visibleRows.length} resources`}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Select
              value={conflictStrategy}
              onValueChange={setConflictStrategy}
            >
              <SelectTrigger className="w-[170px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="FAIL">Conflict: Fail</SelectItem>
                <SelectItem value="OVERWRITE">Conflict: Overwrite</SelectItem>
                <SelectItem value="SKIP">Conflict: Skip</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="primary"
              onClick={() => copySelected()}
              disabled={
                selected.length === 0 ||
                isCopying ||
                info.product !== Product.PRO
              }
            >
              <CopyIcon className="mr-2 h-4 w-4" />
              Copy Selected ({selected.length})
            </Button>
          </div>
        </div>
        <div className="divide-y">
          {!isLoading && visibleRows.length === 0 && (
            <div className="p-6 text-sm text-muted-foreground">
              No resources match the selected filter.
            </div>
          )}
          {visibleRows.map((row) => {
            const canSelect = row.status !== 'target_only';
            return (
              <div key={row.id} className="flex items-start gap-3 p-3">
                <input
                  type="checkbox"
                  className="mt-1 h-4 w-4 rounded border-gray-300"
                  checked={!!selectedRows[row.id]}
                  disabled={!canSelect || isCopying}
                  onChange={(event) =>
                    toggleSelect(row.id, event.target.checked)
                  }
                />
                <div className="min-w-0 flex-1 space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="outlinemuted">{row.resourceType}</Badge>
                    <span className="font-medium">{row.key}</span>
                    {statusBadge(row.status)}
                  </div>
                  {row.differences.length > 0 && (
                    <div className="text-sm text-muted-foreground">
                      Differences: {row.differences.join(', ')}
                    </div>
                  )}
                </div>
                <div className="mt-1">
                  {row.status === 'same' && (
                    <CheckCircle2Icon className="h-4 w-4 text-emerald-600" />
                  )}
                  {row.status === 'different' && (
                    <CircleAlertIcon className="h-4 w-4 text-amber-600" />
                  )}
                  {(row.status === 'source_only' ||
                    row.status === 'target_only') && (
                    <MinusCircleIcon className="h-4 w-4 text-slate-500" />
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {info.product !== Product.PRO && (
        <div className="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-200">
          Read-only compare is available in OSS. Copying selected resources from
          compare view requires Flipt Pro.
        </div>
      )}
    </div>
  );
}
