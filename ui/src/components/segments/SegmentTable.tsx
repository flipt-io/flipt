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
import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { selectSorting, setSorting } from '~/app/segments/segmentsApi';
import { useTimezone } from '~/data/hooks/timezone';
import { ISegment, segmentMatchTypeToLabel } from '~/types/Segment';
import { cn } from '~/lib/utils';
import { Badge } from '~/components/ui/badge';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import { Search } from '~/components/ui/search';
import { DataTableViewOptions } from '~/components/ui/table-view-options';
import Guide from '~/components/ui/guide';
import { useNavigate } from 'react-router-dom';
import { DataTablePagination } from '~/components/ui/table-pagination';

type SegmentTableProps = {
  segments: ISegment[];
};

function Details({ item }: { item: ISegment }) {
  return (
    <div className="flex items-center gap-2 text-xs text-muted-foreground">
      <span>Matches {segmentMatchTypeToLabel(item.matchType)}</span>
      <span className="hidden sm:block">•</span>
      <span className="hidden sm:block">
        Created{' '}
        {formatDistanceToNowStrict(parseISO(item.createdAt), {
          addSuffix: true
        })}
      </span>
      <span>•</span>
      <span>
        Updated{' '}
        {formatDistanceToNowStrict(parseISO(item.updatedAt), {
          addSuffix: true
        })}
      </span>
    </div>
  );
}
export default function SegmentTable(props: SegmentTableProps) {
  const { segments } = props;

  const dispatch = useDispatch();

  const namespace = useSelector(selectCurrentNamespace);
  const { inTimezone } = useTimezone();

  const path = `/namespaces/${namespace.key}/segments`;
  const navigate = useNavigate();

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
  });

  const [filter, setFilter] = useState<string>('');

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
    }),
    columnHelper.accessor((row) => inTimezone(row.createdAt), {
      header: 'Created',
      id: 'createdAt',
      sortingFn: (
        rowA: Row<ISegment>,
        rowB: Row<ISegment>,
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        _columnId: string
      ): number =>
        new Date(rowA.original.createdAt) < new Date(rowB.original.createdAt)
          ? 1
          : -1
    }),
    columnHelper.accessor((row) => inTimezone(row.updatedAt), {
      header: 'Updated',
      id: 'updatedAt',
      sortingFn: (
        rowA: Row<ISegment>,
        rowB: Row<ISegment>,
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        _columnId: string
      ): number =>
        new Date(rowA.original.updatedAt) < new Date(rowB.original.updatedAt)
          ? 1
          : -1
    })
  ];

  const sorting = useSelector(selectSorting);
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

  return (
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between space-x-2">
          <Search
            value={filter ?? ''}
            onChange={setFilter}
            className="h-8 w-[150px] flex-grow text-xs lg:w-[250px]"
          />
          <DataTableViewOptions table={table} />
        </div>
      </div>
      {table.getRowCount() === 0 && (
        <Guide>No segments matched your search.</Guide>
      )}
      {table.getRowModel().rows.map((row) => {
        const item = row.original;
        return (
          <button
            role="link"
            key={row.id}
            className={cn(
              'flex flex-col items-start gap-2 rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent'
            )}
            onClick={() => navigate(`${path}/${item.key}`)}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <div className="font-semibold">{item.name}</div>
                  <Badge variant="outlinemuted" className="hidden sm:block">
                    {item.key}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="line-clamp-2 text-xs text-secondary-foreground">
              {item.description}
            </div>
            <Details item={item} />
          </button>
        );
      })}
      <DataTablePagination table={table} />
    </>
  );
}
