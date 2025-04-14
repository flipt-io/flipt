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
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      <span className="flex items-center gap-1">
        {item.matchType === SegmentMatchType.ALL ? (
          <SigmaIcon className="h-4 w-4" />
        ) : (
          <AsteriskIcon className="h-4 w-4" />
        )}
        Matches {segmentMatchTypeToLabel(item.matchType)}
      </span>
    </div>
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
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between">
          <Searchbox value={filter ?? ''} onChange={setFilter} />
          <DataTableViewOptions table={table} />
        </div>
      </div>
      {table.getRowCount() === 0 && filter.length !== 0 && (
        <Well>
          <p>No segments matched your search.</p>
        </Well>
      )}
      {table.getRowCount() === 0 && filter.length === 0 && (
        <Well>
          <p>
            Segments enable request targeting based on defined criteria. Create
            a{' '}
            <Link className="text-violet-500" to={`${path}/new`}>
              new segment
            </Link>{' '}
            to get started.
          </p>
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
            <SegmentDetails item={item} />
          </button>
        );
      })}
      <DataTablePagination table={table} />
    </>
  );
}
