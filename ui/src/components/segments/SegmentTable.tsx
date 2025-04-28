import {
  PaginationState,
  createColumnHelper,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { AsteriskIcon, SigmaIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router';

import { selectSorting, setSorting } from '~/app/segments/segmentsApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { Button } from '~/components/Button';
import Searchbox from '~/components/Searchbox';
import { DataTablePagination } from '~/components/TablePagination';
import { TableSkeleton } from '~/components/TableSkeleton';
import { DataTableViewOptions } from '~/components/TableViewOptions';
import Well from '~/components/Well';

import { IEnvironment } from '~/types/Environment';
import { INamespace } from '~/types/Namespace';
import {
  ISegment,
  SegmentMatchType,
  segmentMatchTypeToLabel
} from '~/types/Segment';

import { useError } from '~/data/hooks/error';

function SegmentListItem({ item, path }: { item: ISegment; path: string }) {
  const navigate = useNavigate();

  return (
    <button
      role="link"
      className="group w-full rounded-lg border text-left text-sm transition-all hover:bg-accent"
      onClick={() => navigate(path)}
    >
      <div className="flex items-start gap-6 p-4">
        {/* Segment Info and Tags Column */}
        <div className="flex flex-col min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-semibold">{item.name}</span>
            <span className="text-muted-foreground">&middot;</span>
            <code className="text-xs text-muted-foreground font-mono">
              {item.key}
            </code>
          </div>
          {item.description && (
            <p className="mt-1.5 text-sm text-muted-foreground line-clamp-2">
              {item.description}
            </p>
          )}
        </div>

        {/* Status Column */}
        <div className="flex items-center gap-2 shrink-0">
          <div className="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium bg-secondary/50 text-secondary-foreground">
            {item.matchType === SegmentMatchType.ALL ? (
              <SigmaIcon className="h-3.5 w-3.5" />
            ) : (
              <AsteriskIcon className="h-3.5 w-3.5" />
            )}
            Match {segmentMatchTypeToLabel(item.matchType)}
          </div>
        </div>
      </div>
    </button>
  );
}

function EmptySegmentList({ path }: { path: string }) {
  const navigate = useNavigate();

  return (
    <Well>
      <div className="flex flex-col items-center text-center p-4">
        <SigmaIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
        <p className="text-sm text-muted-foreground mb-4">
          Segments enable request targeting based on defined criteria.
        </p>
        <Button
          variant="primary"
          onClick={() => navigate(path)}
          aria-label="New Segment"
        >
          Create Your First Segment
        </Button>
      </div>
    </Well>
  );
}

const columnHelper = createColumnHelper<ISegment>();

const columns = [
  columnHelper.accessor('key', {
    header: 'Key',
    cell: (info) => info.getValue()
  }),
  columnHelper.accessor('name', {
    header: 'Name',
    cell: (info) => info.getValue()
  }),
  columnHelper.accessor('matchType', {
    header: 'Match Type',
    cell: (info) => segmentMatchTypeToLabel(info.getValue())
  }),
  columnHelper.accessor('description', {
    header: 'Description',
    enableSorting: false,
    cell: (info) => info.getValue()
  })
];

type SegmentTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

export default function SegmentTable(props: SegmentTableProps) {
  const { environment, namespace } = props;

  const dispatch = useDispatch();

  const path = `/namespaces/${namespace.key}/segments`;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 25
  });

  const [filter, setFilter] = useState<string>('');

  const sorting = useSelector(selectSorting);

  const { data, isLoading, error } = useListSegmentsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });
  const segments = useMemo(() => data?.segments || [], [data]);
  const hasSegments = segments.length > 0;

  const { setError } = useError();
  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  const table = useReactTable({
    data: segments,
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
            {hasSegments && <DataTableViewOptions table={table} />}
          </div>
        </div>

        {table.getRowCount() === 0 && filter.length === 0 && (
          <EmptySegmentList path={`${path}/new`} />
        )}
        {table.getRowCount() === 0 && filter.length > 0 && (
          <Well>
            <div className="flex flex-col items-center text-center p-4">
              <SigmaIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <p className="text-sm text-muted-foreground">
                No segments matched your search
              </p>
            </div>
          </Well>
        )}

        <div className="space-y-2">
          {table.getRowModel().rows.map((row) => {
            const item = row.original;
            return (
              <SegmentListItem
                key={row.id}
                item={item}
                path={`${path}/${item.key}`}
              />
            );
          })}
        </div>

        {hasSegments && <DataTablePagination table={table} />}
      </div>
    </div>
  );
}
