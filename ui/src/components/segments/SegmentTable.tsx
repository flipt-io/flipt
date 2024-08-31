import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/20/solid';
import {
  createColumnHelper,
  flexRender,
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
import { Link } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { selectSorting, setSorting } from '~/app/segments/segmentsApi';
import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';
import { useTimezone } from '~/data/hooks/timezone';
import { ISegment, segmentMatchTypeToLabel } from '~/types/Segment';
import { truncateKey } from '~/utils/helpers';

type SegmentTableProps = {
  segments: ISegment[];
};

export default function SegmentTable(props: SegmentTableProps) {
  const { segments } = props;

  const dispatch = useDispatch();

  const namespace = useSelector(selectCurrentNamespace);
  const { inTimezone } = useTimezone();

  const path = `/namespaces/${namespace.key}/segments`;

  const searchThreshold = 10;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
  });

  const [filter, setFilter] = useState<string>('');

  const columnHelper = createColumnHelper<ISegment>();

  const columns = [
    columnHelper.accessor('key', {
      header: 'Key',
      cell: (info) => (
        <Link to={`${path}/${info.getValue()}`} className="text-violet-500">
          {truncateKey(info.getValue())}
        </Link>
      ),
      meta: {
        className:
          'truncate whitespace-nowrap py-4 px-3 text-sm font-medium text-gray-900'
      }
    }),
    columnHelper.accessor('name', {
      header: 'Name',
      cell: (info) => info.getValue(),
      meta: {
        className: 'truncate whitespace-nowrap py-4 px-3 text-sm text-gray-500'
      }
    }),
    columnHelper.accessor('matchType', {
      header: 'Match Type',
      cell: (info) => segmentMatchTypeToLabel(info.getValue()),
      meta: {
        className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-600'
      }
    }),
    columnHelper.accessor('description', {
      header: 'Description',
      cell: (info) => info.getValue(),
      meta: {
        className: 'truncate whitespace-nowrap py-4 px-3 text-sm text-gray-500'
      }
    }),
    columnHelper.accessor((row) => inTimezone(row.createdAt), {
      header: 'Created',
      id: 'createdAt',
      meta: {
        className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-500'
      },
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
      meta: {
        className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-500'
      },
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
      {segments.length >= searchThreshold && (
        <Searchbox className="mb-4" value={filter ?? ''} onChange={setFilter} />
      )}
      <table className="divide-y divide-gray-300">
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) =>
                header.column.getCanSort() ? (
                  <th
                    key={header.id}
                    scope="col"
                    className="text-gray-900 px-3 py-3.5 text-left text-sm font-semibold"
                  >
                    <div
                      className="group inline-flex cursor-pointer"
                      onClick={header.column.getToggleSortingHandler()}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                      <span className="text-gray-400 ml-2 flex-none rounded group-hover:visible group-focus:visible">
                        {{
                          asc: (
                            <ChevronUpIcon
                              className="h-5 w-5"
                              aria-hidden="true"
                            />
                          ),
                          desc: (
                            <ChevronDownIcon
                              className="h-5 w-5"
                              aria-hidden="true"
                            />
                          )
                        }[header.column.getIsSorted() as string] ?? null}
                      </span>
                    </div>
                  </th>
                ) : (
                  <th
                    key={header.id}
                    scope="col"
                    className="text-gray-900 px-3 py-3.5 text-left text-sm font-semibold"
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </th>
                )
              )}
            </tr>
          ))}
        </thead>
        <tbody className="divide-y divide-gray-200">
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <td
                  key={cell.id}
                  className={cell.column.columnDef?.meta?.className}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <Pagination
        currentPage={table.getState().pagination.pageIndex + 1}
        totalCount={table.getFilteredRowModel().rows.length}
        pageSize={pagination.pageSize}
        onPageChange={(page: number, size: number) => {
          table.setPageIndex(page - 1);
          table.setPageSize(size);
        }}
      />
    </>
  );
}
