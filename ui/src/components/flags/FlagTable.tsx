import {
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  PaginationState,
  Row,
  useReactTable
} from '@tanstack/react-table';
import { useState, useEffect, useMemo } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { Link, useNavigate } from 'react-router-dom';
import { DataTablePagination } from '~/components/ui/table-pagination';
import { useTimezone } from '~/data/hooks/timezone';
import { FlagType, flagTypeToLabel, IFlag } from '~/types/Flag';
import {
  selectSorting,
  setSorting,
  useListFlagsQuery
} from '~/app/flags/flagsApi';
import { cls } from '~/utils/helpers';
import { Badge } from '~/components/Badge';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import Searchbox from '~/components/Searchbox';
import { DataTableViewOptions } from '~/components/ui/table-view-options';
import { VariableIcon, ToggleLeftIcon } from 'lucide-react';
import { useError } from '~/data/hooks/error';
import { INamespaceBase } from '~/types/Namespace';
import { TableSkeleton } from '~/components/ui/table-skeleton';
import Well from '~/components/Well';

type FlagTableProps = {
  namespace: INamespaceBase;
};

function FlagDetails({ item }: { item: IFlag }) {
  const enabled = item.type === FlagType.BOOLEAN || item.enabled;
  const { inTimezone } = useTimezone();
  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      <Badge variant={enabled ? 'enabled' : 'muted'}>
        {enabled ? 'Enabled' : 'Disabled'}
      </Badge>
      <span className="flex items-center gap-1">
        {item.type === FlagType.BOOLEAN ? (
          <ToggleLeftIcon className="h-4 w-4" />
        ) : (
          <VariableIcon className="h-4 w-4" />
        )}
        {flagTypeToLabel(item.type)}
      </span>
      <span className="hidden sm:block">•</span>
      <time className="hidden sm:block" title={inTimezone(item.createdAt)}>
        Created{' '}
        {formatDistanceToNowStrict(parseISO(item.createdAt), {
          addSuffix: true
        })}
      </time>
      <span>•</span>
      <time title={inTimezone(item.updatedAt)}>
        Updated{' '}
        {formatDistanceToNowStrict(parseISO(item.updatedAt), {
          addSuffix: true
        })}
      </time>
    </div>
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
  }),
  columnHelper.accessor((row) => row.createdAt, {
    header: 'Created',
    id: 'createdAt',
    sortingFn: (
      rowA: Row<IFlag>,
      rowB: Row<IFlag>,
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      _columnId: string
    ): number =>
      new Date(rowA.original.createdAt) < new Date(rowB.original.createdAt)
        ? 1
        : -1
  }),
  columnHelper.accessor((row) => row.updatedAt, {
    header: 'Updated',
    id: 'updatedAt',
    sortingFn: (
      rowA: Row<IFlag>,
      rowB: Row<IFlag>,
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      _columnId: string
    ): number =>
      new Date(rowA.original.updatedAt) < new Date(rowB.original.updatedAt)
        ? 1
        : -1
  })
];

export default function FlagTable(props: FlagTableProps) {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const namespace = props.namespace;

  const path = `/namespaces/${namespace.key}/flags`;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
  });

  const [filter, setFilter] = useState<string>('');

  const sorting = useSelector(selectSorting);

  const { data, isLoading, error } = useListFlagsQuery(namespace.key);
  const flags = useMemo(() => data?.flags || [], [data]);

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
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between">
          <Searchbox value={filter ?? ''} onChange={setFilter} />
          <DataTableViewOptions table={table} />
        </div>
      </div>
      {table.getRowCount() === 0 && filter.length === 0 && (
        <Well>
          <p>
            Flags enable you to control and roll out new functionality
            dynamically. Create a{' '}
            <Link className="text-violet-500" to={`${path}/new`}>
              new flag
            </Link>{' '}
            to get started.
          </p>
        </Well>
      )}
      {table.getRowCount() === 0 && filter.length > 0 && (
        <Well>
          <p>No flags matched your search.</p>
        </Well>
      )}
      {table.getRowModel().rows.map((row) => {
        const item = row.original;
        return (
          <button
            role="link"
            key={row.id}
            className={cls(
              'flex flex-col items-start gap-2 rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent'
            )}
            onClick={() => navigate(`${path}/${item.key}`)}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <div className="truncate font-semibold">{item.name}</div>
                  <Badge variant="outlinemuted" className="hidden sm:block">
                    {item.key}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="line-clamp-2 text-xs text-secondary-foreground">
              {item.description}
            </div>
            <FlagDetails item={item} />
          </button>
        );
      })}
      <DataTablePagination table={table} />
    </>
  );
}
