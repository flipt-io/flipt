import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { ToggleLeftIcon, VariableIcon } from 'lucide-react';
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
import { cls } from '~/utils/helpers';

type FlagTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

export function FlagBadge({ item }: { item: IFlag }) {
  return (
    <span className="flex items-center gap-1 text-muted-foreground">
      {item.type === FlagType.BOOLEAN ? (
        <ToggleLeftIcon className="h-4 w-4" />
      ) : (
        <VariableIcon className="h-4 w-4" />
      )}
      {flagTypeToLabel(item.type)}
    </span>
  );
}

export function FlagDetails({ item }: { item: IFlag }) {
  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      {item.type === FlagType.BOOLEAN ? (
        <Badge variant={item.enabled ? 'enabled' : 'destructiveoutline'}>
          {item.enabled ? 'True' : 'False'}
        </Badge>
      ) : (
        <Badge variant={item.enabled ? 'enabled' : 'muted'}>
          {item.enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      )}
      <FlagBadge item={item} />
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
  })
];

export default function FlagTable(props: FlagTableProps) {
  const { environment, namespace } = props;

  const navigate = useNavigate();
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
              <div className="flex w-full items-center justify-between">
                <div className="flex max-w-[70%] items-center gap-2">
                  <div className="truncate font-semibold">{item.name}</div>
                </div>
                <Badge
                  variant="outlinemuted"
                  className="ml-2 hidden flex-shrink-0 sm:flex"
                >
                  {item.key}
                </Badge>
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
