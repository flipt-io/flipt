import Box from '@mui/material/Box';
import { SparkLineChart } from '@mui/x-charts';
import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import {
  CheckSquareIcon,
  FlagIcon,
  PowerIcon,
  ToggleLeftIcon,
  VariableIcon,
  XIcon,
  XSquareIcon
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { useGetBatchFlagEvaluationCountQuery } from '~/app/flags/analyticsApi';
import {
  selectSorting,
  setSorting,
  useListFlagsQuery
} from '~/app/flags/flagsApi';
import { selectInfo } from '~/app/meta/metaSlice';

import { Badge } from '~/components/Badge';
import { Button } from '~/components/Button';
import Searchbox from '~/components/Searchbox';
import { DataTablePagination } from '~/components/TablePagination';
import { TableSkeleton } from '~/components/TableSkeleton';
import { DataTableViewOptions } from '~/components/TableViewOptions';
import Well from '~/components/Well';
import MetadataFilterPopover from '~/components/flags/MetadataFilterPopover';

import { IBatchFlagEvaluationCount } from '~/types/Analytics';
import { IEnvironment } from '~/types/Environment';
import { FlagType, IFlag, MetadataFilter, flagTypeToLabel } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { applyMetadataFilters } from '~/utils/flagMetadataFilter';
import { cls } from '~/utils/helpers';

function VariantFlagBadge({ enabled }: { enabled: boolean }) {
  return (
    <div
      className={cls(
        'inline-flex items-center gap-1.5 px-2 py-1 text-xs font-medium transition-colors rounded-md',
        enabled
          ? 'text-green-700 bg-green-100/40 dark:bg-green-600/20'
          : 'text-orange-700 bg-orange-100/40 dark:bg-orange-600/20'
      )}
    >
      <PowerIcon className="h-3.5 w-3.5" />
      {enabled ? 'Enabled' : 'Disabled'}
    </div>
  );
}

function BooleanFlagBadge({ enabled }: { enabled: boolean }) {
  return (
    <div
      title="The default value for this flag"
      className={cls(
        'inline-flex items-center gap-1.5 px-2 py-1 text-xs font-medium transition-colors rounded-md',
        enabled
          ? 'text-green-700 bg-green-100/40 dark:bg-green-600/20'
          : 'text-orange-700 bg-orange-100/40 dark:bg-orange-600/20'
      )}
    >
      {enabled ? (
        <>
          <CheckSquareIcon className="h-3.5 w-3.5" />
          True
        </>
      ) : (
        <>
          <XSquareIcon className="h-3.5 w-3.5" />
          False
        </>
      )}
    </div>
  );
}

function CombinedFlagBadge({ item }: { item: IFlag }) {
  return (
    <div className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium bg-secondary/50 text-muted-foreground">
      {item.type === FlagType.BOOLEAN ? (
        <ToggleLeftIcon className="h-3.5 w-3.5" />
      ) : (
        <VariableIcon className="h-3.5 w-3.5" />
      )}
      {flagTypeToLabel(item.type)}
    </div>
  );
}

function FlagListItem({
  item,
  path,
  evaluationValues = []
}: {
  item: IFlag;
  path: string;
  evaluationValues?: number[];
}) {
  const navigate = useNavigate();

  return (
    <button
      role="link"
      aria-label={item.key}
      className="group w-full rounded-lg border text-left text-sm transition-all hover:bg-accent"
      onClick={() => navigate(path)}
    >
      <div className="flex items-start gap-6 p-4">
        {/* Flag Info and Tags Column */}
        <div className="flex flex-col min-w-0 flex-1">
          <div className="flex items-center gap-3">
            <span className="font-semibold text-base">{item.name}</span>
            <Badge variant="outlinemuted">{item.key}</Badge>
          </div>
          {item.description && (
            <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}
          <div className="flex gap-3 mt-4">
            <CombinedFlagBadge item={item} />
            {/* Status Column - Contains both type badge and toggle */}
            {item.type === FlagType.VARIANT ? (
              <VariantFlagBadge enabled={item.enabled} />
            ) : (
              <BooleanFlagBadge enabled={item.enabled} />
            )}
          </div>
        </div>

        <div className="pl-6 shrink-0 hidden sm:block">
          {/* Sparkline */}
          <div className="mt-2 flex items-center min-h-[24px]">
            {evaluationValues.length > 0 && (
              <div
                className="w-[256px]"
                title="Evaluation requests in the last 24 hours"
              >
                <Box sx={{ flexGrow: 1 }}>
                  <SparkLineChart
                    data={evaluationValues}
                    color="var(--brand)"
                    curve="natural"
                    height={24}
                    yAxis={{
                      min: 0,
                      max: Math.max(...evaluationValues) + 1
                    }}
                  />
                </Box>
              </div>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}

function EmptyFlagList({ path }: { path: string }) {
  const navigate = useNavigate();

  return (
    <Well>
      <div className="flex flex-col items-center text-center p-4">
        <FlagIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
        <p className="text-sm text-muted-foreground mb-4">
          Flags enable you to control and roll out new functionality
          dynamically.
        </p>
        <Button
          aria-label="New Flag"
          variant="primary"
          onClick={() => navigate(path)}
        >
          Create Your First Flag
        </Button>
      </div>
    </Well>
  );
}

const columnHelper = createColumnHelper<
  IFlag & Partial<IBatchFlagEvaluationCount>
>();

const columns = [
  columnHelper.accessor('key', {
    header: 'Key',
    cell: (info) => info.getValue()
  }),
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (info) => info.getValue()
  }),
  columnHelper.accessor('description', {
    header: 'Description',
    enableSorting: false,
    cell: (info) => info.getValue()
  }),
  columnHelper.accessor((row) => row, {
    header: 'Evaluation',
    enableSorting: false,
    cell: (info) => (info.getValue() ? 'Enabled' : 'Disabled')
  }),
  columnHelper.accessor('type', {
    header: 'Type',
    cell: (info) => flagTypeToLabel(info.getValue())
  }),
  columnHelper.accessor('key', {
    id: 'sparkline',
    header: 'Evaluations',
    enableSorting: false,
    cell: (info) => {
      const flagKey = info.getValue();
      const values = info.row.original.flagEvaluations?.[flagKey]?.values ?? [];
      if (!values.length) {
        return <span className="text-muted-foreground">No data</span>;
      }
      return <span></span>;
    }
  })
];

type FlagTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

export default function FlagTable(props: FlagTableProps) {
  const { environment, namespace } = props;

  const dispatch = useDispatch();

  const path = `/namespaces/${namespace.key}/flags`;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 25
  });

  const [filter, setFilter] = useState<string>('');

  const sorting = useSelector(selectSorting);

  const { data, isLoading, error } = useListFlagsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  const info = useSelector(selectInfo);
  const flags = useMemo(() => data?.flags || [], [data]);
  const flagKeys = useMemo(() => flags.map((f) => f.key), [flags]);
  const hasFlags = flags.length > 0;

  const [metadataFilters, setMetadataFilters] = useState<MetadataFilter[]>([]);

  const availableMetadataKeys = useMemo(
    () =>
      [...new Set(flags.flatMap((f) => Object.keys(f.metadata ?? {})))].sort(),
    [flags]
  );

  const filteredFlags = useMemo(
    () => applyMetadataFilters(flags, metadataFilters),
    [flags, metadataFilters]
  );

  const { setError } = useError();
  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  const table = useReactTable({
    data: filteredFlags,
    columns,
    state: {
      globalFilter: filter,
      sorting,
      pagination
    },
    globalFilterFn: 'includesString',
    onSortingChange: (updater) => {
      const newSorting =
        typeof updater === 'function' ? updater(sorting) : updater;
      dispatch(setSorting(newSorting));
    },
    onPaginationChange: setPagination,
    onGlobalFilterChange: setFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel()
  });

  // TODO: this gets all the flag evaluations for all flags, we should only fetch the ones that are currently visible
  // we'll likely need to switch to a server side pagination model to get this to perform well
  const { data: evaluationCount } = useGetBatchFlagEvaluationCountQuery(
    {
      environmentKey: environment.key,
      namespaceKey: namespace.key,
      flagKeys
    },
    {
      skip: !info.analytics?.enabled || isLoading
    }
  );

  if (isLoading) {
    return <TableSkeleton />;
  }

  return (
    <div className="w-full">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex flex-1 items-center justify-between">
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
              <MetadataFilterPopover
                availableKeys={availableMetadataKeys}
                onAdd={(f) => setMetadataFilters((prev) => [...prev, f])}
              />
            </div>
            {hasFlags && <DataTableViewOptions table={table} />}
          </div>
        </div>

        {metadataFilters.length > 0 && (
          <div className="flex flex-wrap items-center gap-2">
            {metadataFilters.map((f, i) => (
              <Badge
                key={i}
                variant="outlinemuted"
                className="flex items-center gap-1 pr-1"
              >
                <span>
                  {f.key}: {f.value}
                </span>
                <button
                  type="button"
                  onClick={() =>
                    setMetadataFilters((prev) =>
                      prev.filter((_, idx) => idx !== i)
                    )
                  }
                  aria-label={`Remove filter ${f.key}:${f.value}`}
                  className="ml-1 rounded-full hover:bg-muted p-0.5"
                >
                  <XIcon className="h-3 w-3" />
                </button>
              </Badge>
            ))}
            <button
              type="button"
              onClick={() => setMetadataFilters([])}
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              Clear all
            </button>
          </div>
        )}

        {table.getRowCount() === 0 &&
          filter.length === 0 &&
          metadataFilters.length === 0 && (
            <EmptyFlagList path={`${path}/new`} />
          )}
        {table.getRowCount() === 0 &&
          (filter.length > 0 || metadataFilters.length > 0) && (
            <Well>
              <div className="flex flex-col items-center text-center p-4">
                <FlagIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
                <p className="text-sm text-muted-foreground">
                  No flags matched your search
                </p>
              </div>
            </Well>
          )}

        <div className="space-y-2">
          {table.getRowModel().rows.map((row) => {
            const item = row.original;
            const values =
              evaluationCount?.flagEvaluations?.[item.key]?.values ?? [];
            return (
              <FlagListItem
                key={row.id}
                item={item}
                path={`${path}/${item.key}`}
                evaluationValues={values}
              />
            );
          })}
        </div>

        {hasFlags && <DataTablePagination table={table} />}
      </div>
    </div>
  );
}
