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
  SortingState,
  useReactTable
} from '@tanstack/react-table';
import { formatDistanceToNowStrict, parseISO } from 'date-fns';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';
import { IFlag } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';
import { truncateKey } from '~/utils/helpers';

type FlagTableProps = {
  namespace: INamespace;
  flags: IFlag[];
};

export default function FlagTable(props: FlagTableProps) {
  const { namespace, flags } = props;

  const path = `/namespaces/${namespace.key}/flags`;

  const pageSize = 20;
  const searchThreshold = 10;

  const [sorting, setSorting] = useState<SortingState>([]);
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize
  });

  const [filter, setFilter] = useState<string>('');

  const columnHelper = createColumnHelper<IFlag>();

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
    columnHelper.accessor('enabled', {
      header: 'Status',
      cell: (info) => (
        <span
          className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold leading-5 ${
            info.getValue()
              ? 'bg-green-100 text-green-600'
              : 'bg-gray-100 text-gray-500'
          }`}
        >
          {info.getValue() ? 'Enabled' : 'Disabled'}
        </span>
      ),
      meta: {
        className: 'whitespace-nowrap py-4 px-3 text-sm'
      }
    }),
    columnHelper.accessor('description', {
      header: 'Description',
      cell: (info) => info.getValue(),
      meta: {
        className: 'truncate whitespace-nowrap py-4 px-3 text-sm text-gray-500'
      }
    }),
    columnHelper.accessor(
      (row) => formatDistanceToNowStrict(parseISO(row.createdAt)),
      {
        header: 'Created',
        id: 'createdAt',
        meta: {
          className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-500'
        },
        sortingFn: (
          rowA: Row<IFlag>,
          rowB: Row<IFlag>,
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          _columnId: string
        ): number =>
          new Date(rowA.original.createdAt) < new Date(rowB.original.createdAt)
            ? 1
            : -1
      }
    ),
    columnHelper.accessor(
      (row) => formatDistanceToNowStrict(parseISO(row.updatedAt)),
      {
        header: 'Updated',
        id: 'updatedAt',
        meta: {
          className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-500'
        },
        sortingFn: (
          rowA: Row<IFlag>,
          rowB: Row<IFlag>,
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          _columnId: string
        ): number =>
          new Date(rowA.original.updatedAt) < new Date(rowB.original.updatedAt)
            ? 1
            : -1
      }
    )
  ];

  const table = useReactTable({
    data: flags,
    columns,
    state: {
      globalFilter: filter,
      sorting,
      pagination
    },
    globalFilterFn: 'includesString',
    onSortingChange: setSorting,
    onPaginationChange: setPagination,
    onGlobalFilterChange: setFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel()
  });

  return (
    <>
      {flags.length >= searchThreshold && (
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
                    className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
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
                      <span className="ml-2 flex-none rounded text-gray-400 group-hover:visible group-focus:visible">
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
                    className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
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
      {table.getPageCount() > 1 && (
        <Pagination
          className="mt-4"
          currentPage={table.getState().pagination.pageIndex + 1}
          totalCount={table.getFilteredRowModel().rows.length}
          pageSize={pageSize}
          onPageChange={(page: number) => {
            table.setPageIndex(page - 1);
          }}
        />
      )}
    </>
  );
}
