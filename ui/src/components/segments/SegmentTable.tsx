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
import { Link, useNavigate } from 'react-router';

import { selectSorting, setSorting } from '~/app/segments/segmentsApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { Badge } from '~/components/Badge';
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
import { cls } from '~/utils/helpers';

type SegmentTableProps = {
  environment: IEnvironment;
  namespace: INamespace;
};

function SegmentDetails({ item }: { item: ISegment }) {
  return (
    <div className="flex items-center gap-2">
      <Badge variant="outlinemuted" className="flex items-center gap-1">
        {item.matchType === SegmentMatchType.ALL ? (
          <SigmaIcon className="h-4 w-4" />
        ) : (
          <AsteriskIcon className="h-4 w-4" />
        )}
        {segmentMatchTypeToLabel(item.matchType)}
      </Badge>
      {item.constraints && item.constraints.length > 0 && (
        <Badge variant="outlinemuted">
          {item.constraints.length} constraint
          {item.constraints.length !== 1 ? 's' : ''}
        </Badge>
      )}
    </div>
  );
}

function SegmentListItem({
  item,
  onClick
}: {
  item: ISegment;
  onClick: () => void;
}) {
  return (
    <button
      role="link"
      className="flex w-full items-center justify-between rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent"
      onClick={onClick}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate font-semibold">{item.name}</span>
        </div>
        <code className="text-xs text-muted-foreground font-mono">
          {item.key}
        </code>
        {item.description && (
          <p className="mt-3 line-clamp-2 text-sm text-muted-foreground">
            {item.description}
          </p>
        )}
      </div>
      <div className="flex items-center gap-2 pl-3">
        <SegmentDetails item={item} />
      </div>
    </button>
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

function EmptySegmentList({ path }: { path: string }) {
  return (
    <Well>
      <div className="flex flex-col items-center text-center p-4">
        <SigmaIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
        <p className="text-sm text-muted-foreground mb-4">
          Segments enable request targeting based on defined criteria.
        </p>
        <Link
          className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-violet-500 text-white hover:bg-violet-600 h-9 px-4 py-2"
          to={`${path}/new`}
        >
          Create Your First Segment
        </Link>
      </div>
    </Well>
  );
}

export default function SegmentTable(props: SegmentTableProps) {
  const { environment, namespace } = props;

  const navigate = useNavigate();
  const dispatch = useDispatch();

  const path = `/namespaces/${namespace.key}/segments`;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
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
    <div className={cls(hasSegments ? 'w-full' : '')}>
      <div className={cls('space-y-4')}>
        <div className="flex items-center justify-between">
          <div className="flex flex-1 items-center justify-between">
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
            </div>
            {hasSegments && <DataTableViewOptions table={table} />}
          </div>
        </div>

        {table.getRowCount() === 0 && filter.length === 0 && (
          <EmptySegmentList path={path} />
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
                onClick={() => navigate(`${path}/${item.key}`)}
              />
            );
          })}
        </div>

        {hasSegments && <DataTablePagination table={table} />}
      </div>
    </div>
  );
}
