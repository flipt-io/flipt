import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { FlagIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router';

import {
  selectSorting,
  setSorting,
  useListFlagsQuery
} from '~/app/flags/flagsApi';

import { Badge } from '~/components/Badge';
import Searchbox from '~/components/Searchbox';
import { DataTablePagination } from '~/components/TablePagination';
import { TableSkeleton } from '~/components/TableSkeleton';
import { DataTableViewOptions } from '~/components/TableViewOptions';
import Well from '~/components/Well';

import { IEnvironment } from '~/types/Environment';
import { FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';

import { FlagTypeBadge } from './FlagTypeBadge';

function FlagDetails({ item }: { item: IFlag }) {
  return (
    <div className="flex items-center gap-2">
      <FlagTypeBadge type={item.type} />
      {item.type === FlagType.BOOLEAN ? (
        <Badge variant="secondary" state={item.enabled ? 'success' : 'error'}>
          Default: {item.enabled ? 'True' : 'False'}
        </Badge>
      ) : (
        <Badge variant="secondary" state={item.enabled ? 'none' : 'muted'}>
          Status: {item.enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      )}
    </div>
  );
}

function FlagListItem({ item }: { item: IFlag & { namespace: string } }) {
  const navigate = useNavigate();

  return (
    <div className="flex gap-2">
      <button
        role="link"
        className="flex w-full items-center justify-between rounded-lg border p-5 text-left text-sm transition-all hover:bg-accent"
        onClick={() =>
          navigate(`/namespaces/${item.namespace}/flags/${item.key}`)
        }
      >
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate font-semibold">{item.name}</span>
          </div>
          <code className="text-xs text-muted-foreground font-mono">
            {item.key}
          </code>
          {item.description && (
            <p className="mt-3 text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}
        </div>
        <div className="flex items-center gap-2 pl-3">
          <FlagDetails item={item} />
        </div>
      </button>
    </div>
  );
}

function EmptyFlagList({ path }: { path: string }) {
  return (
    <Well>
      <div className="flex flex-col items-center text-center p-4">
        <FlagIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
        <p className="text-sm text-muted-foreground mb-4">
          Flags enable you to control and roll out new functionality
          dynamically.
        </p>
        <Link
          className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-violet-500 text-white hover:bg-violet-600 h-9 px-4 py-2"
          to={`${path}/new`}
        >
          Create Your First Flag
        </Link>
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
    pageSize: 20
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
    <div className={'flex gap-6'}>
      {/* Flag List */}
      <div className={'w-full space-y-4'}>
        <div className="flex items-center justify-between">
          <div className="flex flex-1 items-center justify-between">
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
            </div>
            {hasFlags && <DataTableViewOptions table={table} />}
          </div>
        </div>

        {table.getRowCount() === 0 && filter.length === 0 && (
          <EmptyFlagList path={path} />
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
                item={item as IFlag & { namespace: string }}
              />
            );
          })}
        </div>

        {hasFlags && <DataTablePagination table={table} />}
      </div>
    </div>
  );
}
