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
  VariableIcon
} from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import {
  selectSorting,
  setSorting,
  useListFlagsQuery
} from '~/app/flags/flagsApi';

import { Button } from '~/components/Button';
import Searchbox from '~/components/Searchbox';
import { DataTablePagination } from '~/components/TablePagination';
import { TableSkeleton } from '~/components/TableSkeleton';
import { DataTableViewOptions } from '~/components/TableViewOptions';
import Well from '~/components/Well';

import { IEnvironment } from '~/types/Environment';
import { FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { cls } from '~/utils/helpers';

function VariantFlagBadge({ enabled }: { enabled: boolean }) {
  return (
    <div
      className={cls(
        'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
        enabled
          ? 'bg-green-100/50 text-green-700 dark:bg-green-500/10 dark:text-green-400'
          : 'bg-red-100/50 text-red-700 dark:bg-red-500/10 dark:text-red-400'
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
      className={cls(
        'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
        enabled
          ? 'bg-green-100/50 text-green-700 dark:bg-green-500/10 dark:text-green-400'
          : 'bg-red-100/50 text-red-700 dark:bg-red-500/10 dark:text-red-400'
      )}
    >
      <CheckSquareIcon className="h-3.5 w-3.5" />
      Default {enabled ? 'True' : 'False'}
    </div>
  );
}

function CombinedFlagBadge({ item }: { item: IFlag }) {
  return (
    <div className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium bg-secondary/50 text-secondary-foreground">
      {item.type === FlagType.BOOLEAN ? (
        <ToggleLeftIcon className="h-3.5 w-3.5" />
      ) : (
        <VariableIcon className="h-3.5 w-3.5" />
      )}
      {flagTypeToLabel(item.type)}
    </div>
  );
}

function FlagListItem({ item, path }: { item: IFlag; path: string }) {
  const navigate = useNavigate();

  return (
    <button
      role="link"
      className="group w-full rounded-lg border text-left text-sm transition-all hover:bg-accent"
      onClick={() => navigate(path)}
    >
      <div className="flex items-start gap-6 p-4">
        {/* Flag Info and Tags Column */}
        <div className="flex flex-col min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-base">{item.name}</span>
            <span className="text-muted-foreground">&middot;</span>
            <code className="text-xs text-muted-foreground font-mono">
              {item.key}
            </code>
          </div>
          {item.description && (
            <p className="mt-2 text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}
        </div>

        {/* Status Column - Contains both type badge and toggle */}
        <div className="flex items-center gap-3 pl-6 shrink-0">
          <CombinedFlagBadge item={item} />
          {item.type === FlagType.VARIANT ? (
            <VariantFlagBadge enabled={item.enabled} />
          ) : (
            <BooleanFlagBadge enabled={item.enabled} />
          )}
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
        <Button variant="primary" onClick={() => navigate(path)}>
          Create Your First Flag
        </Button>
      </div>
    </Well>
  );
}

const columnHelper = createColumnHelper<IFlag>();

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
  const flags = useMemo(() => data?.flags || [], [data]);
  const hasFlags = flags.length > 0;

  const { setError } = useError();
  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  const table = useReactTable({
    data: flags,
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
            </div>
            {hasFlags && <DataTableViewOptions table={table} />}
          </div>
        </div>

        {table.getRowCount() === 0 && filter.length === 0 && (
          <EmptyFlagList path={`${path}/new`} />
        )}
        {table.getRowCount() === 0 && filter.length > 0 && (
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
            return (
              <FlagListItem
                key={row.id}
                item={item}
                path={`${path}/${item.key}`}
              />
            );
          })}
        </div>

        {hasFlags && <DataTablePagination table={table} />}
      </div>
    </div>
  );
}
