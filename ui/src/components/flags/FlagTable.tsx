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
import { useSelector, useDispatch } from 'react-redux';
import { Link } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';
import { useTimezone } from '~/data/hooks/timezone';
import { FlagType, flagTypeToLabel, IFlag } from '~/types/Flag';
import { truncateKey } from '~/utils/helpers';
import { selectSorting, setSorting } from '~/app/flags/flagsApi';

function flagDefaultValue(flag: IFlag): string {
  switch (flag.type) {
    case FlagType.BOOLEAN:
      return flag.enabled ? 'True' : 'False';
    case FlagType.VARIANT:
      return flag.defaultVariant?.key || '';
    default:
      return '';
  }
}

type FlagTableProps = {
  flags: IFlag[];
};

export default function FlagTable(props: FlagTableProps) {
  const { flags } = props;

  const dispatch = useDispatch();

  const namespace = useSelector(selectCurrentNamespace);
  const { inTimezone } = useTimezone();

  const path = `/namespaces/${namespace.key}/flags`;

  const searchThreshold = 10;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
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
    columnHelper.accessor(
      (row) => row.type === FlagType.BOOLEAN || row.enabled,
      {
        header: 'Evaluation',
        cell: (info) => (
          <span
            className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold leading-5 ${
              info.getValue()
                ? 'text-green-600 bg-green-100'
                : 'text-gray-500 bg-gray-100'
            }`}
          >
            {info.getValue() ? 'Enabled' : 'Disabled'}
          </span>
        ),
        meta: {
          className: 'whitespace-nowrap py-4 px-3 text-sm'
        }
      }
    ),
    columnHelper.accessor('type', {
      header: 'Type',
      cell: (info) => flagTypeToLabel(info.getValue()),
      meta: {
        className: 'whitespace-nowrap py-4 px-3 text-sm text-gray-600'
      }
    }),
    columnHelper.accessor((row) => row, {
      header: 'Default Value',
      cell: (info) => flagDefaultValue(info.getValue()),
      meta: {
        className: 'truncate whitespace-nowrap py-4 px-3 text-sm text-gray-600'
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
        rowA: Row<IFlag>,
        rowB: Row<IFlag>,
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

  const sorting = useSelector(selectSorting);
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
