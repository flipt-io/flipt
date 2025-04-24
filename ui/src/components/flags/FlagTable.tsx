import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { ChevronDownIcon, FlagIcon } from 'lucide-react';
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
import { JsonEditor } from '~/components/json/JsonEditor';

import { IEnvironment } from '~/types/Environment';
import { FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { cls } from '~/utils/helpers';

import { FlagTypeBadge } from './FlagTypeBadge';

type FlagTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

function FlagListItem({
  item,
  isSelected,
  onClick
}: {
  item: IFlag;
  isSelected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      role="link"
      className={cls(
        'flex w-full items-center justify-between rounded-lg border p-3 text-left text-sm transition-all',
        {
          'border-violet-500 bg-violet-50 dark:bg-violet-900/10': isSelected,
          'hover:bg-accent': !isSelected
        }
      )}
      onClick={onClick}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate font-semibold">{item.name}</span>
          <FlagTypeBadge type={item.type} />
        </div>
        <code className="text-xs text-muted-foreground font-mono">
          {item.key}
        </code>
      </div>
      <div className="flex items-center gap-2 pl-3">
        <FlagDetails item={item} />
      </div>
    </button>
  );
}

function FlagDetails({ item }: { item: IFlag }) {
  if (item.type === FlagType.BOOLEAN) {
    // For boolean flags, show the actual value (true/false)
    return (
      <div className="flex items-center gap-2">
        <Badge variant={item.enabled ? 'enabled' : 'destructive'}>
          {item.enabled ? 'True' : 'False'}
        </Badge>
      </div>
    );
  }

  // For variant flags, show enabled/disabled state
  return (
    <div className="flex items-center gap-2">
      <Badge variant={item.enabled ? 'outline' : 'muted'}>
        {item.enabled ? 'Enabled' : 'Disabled'}
      </Badge>
    </div>
  );
}

function FlagDetailPanel({
  flag,
  namespace
}: {
  flag: IFlag;
  namespace: string;
}) {
  const navigate = useNavigate();
  const [isMetadataExpanded, setIsMetadataExpanded] = useState(true);

  return (
    <div className="space-y-6 rounded-lg border p-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">{flag.name}</h2>
          <code className="text-sm text-muted-foreground font-mono">
            {flag.key}
          </code>
        </div>
        <div className="flex items-center gap-2">
          <FlagTypeBadge type={flag.type} />
          <FlagDetails item={flag} />
        </div>
      </div>

      {flag.description && (
        <div>
          <h3 className="text-sm font-medium text-secondary-foreground mb-1">
            Description
          </h3>
          <p className="text-sm text-secondary-foreground">
            {flag.description}
          </p>
        </div>
      )}

      <div>
        <button
          onClick={() => setIsMetadataExpanded(!isMetadataExpanded)}
          className="flex w-full items-center justify-between mb-1"
        >
          <h3 className="text-sm font-medium text-secondary-foreground">
            Metadata
          </h3>
          <ChevronDownIcon
            className={cls(
              'h-4 w-4 text-secondary-foreground transition-transform duration-200',
              {
                'transform rotate-180': isMetadataExpanded
              }
            )}
          />
        </button>
        {isMetadataExpanded && (
          <div className="rounded-md border bg-muted/50">
            <JsonEditor
              id="flag-metadata-viewer"
              value={JSON.stringify(flag.metadata, null, 2)}
              setValue={() => {}} // No-op since this is read-only
              disabled={true}
              strict={false}
              height="20vh"
            />
          </div>
        )}
      </div>

      {/* Placeholder for evaluation analytics */}
      {/* {flag.enabled && (
        <div>
          <h3 className="text-sm font-medium text-secondary-foreground mb-1">Evaluation Analytics</h3>
          <div className="text-sm text-muted-foreground">Analytics coming soon...</div>
        </div>
      )} */}

      <div className="flex justify-end">
        <button
          className="text-sm text-violet-500 hover:text-violet-600"
          onClick={() => navigate(`/namespaces/${namespace}/flags/${flag.key}`)}
        >
          Edit →
        </button>
      </div>
    </div>
  );
}

function EmptyFlagDetails() {
  return (
    <div className="flex h-full flex-col items-center justify-center text-center p-6">
      <FlagIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
      <h3 className="text-lg font-medium text-muted-foreground mb-2">
        No Flag Selected
      </h3>
      <p className="text-sm text-muted-foreground">
        Select a flag from the list to view its details
      </p>
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

  const hasFlags = table.getRowCount() > 0;

  return (
    <div className={cls(hasFlags ? 'flex gap-6' : '')}>
      {/* Flag List */}
      <div className={cls(hasFlags ? 'w-1/2' : 'w-full', 'space-y-4')}>
        <div className="flex items-center justify-between">
          <div className="flex flex-1 items-center justify-between">
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
              {hasFlags && (
                <Link
                  to={`${path}/new`}
                  className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-violet-500 text-white hover:bg-violet-600 h-9 px-4 py-2"
                >
                  + New Flag
                </Link>
              )}
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
