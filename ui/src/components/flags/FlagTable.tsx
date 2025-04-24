import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { ChevronRightIcon, FlagIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link, useNavigate } from 'react-router';

import {
  selectSorting,
  setSorting,
  useListFlagsQuery
} from '~/app/flags/flagsApi';

import Searchbox from '~/components/Searchbox';
import { DataTablePagination } from '~/components/TablePagination';
import { TableSkeleton } from '~/components/TableSkeleton';
import { DataTableViewOptions } from '~/components/TableViewOptions';
import Well from '~/components/Well';

import { IEnvironment } from '~/types/Environment';
import { IFlag, flagTypeToLabel } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { cls } from '~/utils/helpers';

import { EmptyFlagDetails, FlagDetailPanel, FlagDetails } from './FlagDetails';
import { FlagTypeBadge } from './FlagTypeBadge';

type FlagTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

function FlagListItem({
  item,
  isSelected,
  onClick,
  namespace
}: {
  item: IFlag;
  isSelected: boolean;
  onClick: () => void;
  namespace: string;
}) {
  const navigate = useNavigate();

  // Stop event propagation when clicking details button
  const handleDetailsClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onClick();
  };

  return (
    <div className="flex gap-2">
      <button
        role="link"
        className={cls(
          'flex w-full items-center justify-between rounded-lg border p-3 text-left text-sm transition-all',
          {
            'border-violet-500 bg-violet-50 dark:bg-violet-900/10': isSelected,
            'hover:bg-accent': !isSelected
          }
        )}
        onClick={() => navigate(`/namespaces/${namespace}/flags/${item.key}`)}
      >
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate font-semibold">{item.name}</span>
            <FlagTypeBadge type={item.type} />
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
      <button
        className="flex items-center justify-center rounded-lg p-2 transition-colors hover:bg-accent"
        onClick={handleDetailsClick}
        title="View details"
      >
        <ChevronRightIcon className="h-4 w-4 text-muted-foreground" />
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

export default function FlagTable(props: FlagTableProps) {
  const { environment, namespace } = props;
  const [selectedFlag, setSelectedFlag] = useState<IFlag | null>(null);

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

  const hasFlags = table.getRowCount() > 0;

  return (
    <div className={cls(hasFlags ? 'flex gap-6' : '')}>
      {/* Flag List */}
      <div className={cls(hasFlags ? 'w-1/2' : 'w-full', 'space-y-4')}>
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
                item={item}
                isSelected={selectedFlag?.key === item.key}
                onClick={() => setSelectedFlag(item)}
                namespace={namespace.key}
              />
            );
          })}
        </div>

        {hasFlags && <DataTablePagination table={table} />}
      </div>

      {/* Flag Details - Only show when there are flags */}
      {hasFlags && (
        <div className="w-1/2">
          {selectedFlag ? (
            <FlagDetailPanel flag={selectedFlag} namespace={namespace.key} />
          ) : (
            <div className="h-[400px] rounded-lg border">
              <EmptyFlagDetails />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
