import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  PaginationState,
  Row,
  RowSelectionState,
  SortingState,
  Table,
  useReactTable
} from '@tanstack/react-table';
import { format, parseISO } from 'date-fns';
import React, { HTMLProps, useEffect, useState } from 'react';
import Button from '~/components/forms/buttons/Button';
import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';
import { IAuthToken } from '~/types/auth/Token';

type TokenRowActionsProps = {
  row: Row<IAuthToken>;
  setDeletingTokens: (tokens: IAuthToken[]) => void;
  setShowDeleteTokenModal: (show: boolean) => void;
};

function TokenRowActions(props: TokenRowActionsProps) {
  const { row, setDeletingTokens, setShowDeleteTokenModal } = props;

  let className = 'text-violet-600 hover:text-violet-900';
  if (row.getIsSelected()) {
    className = 'text-gray-400 hover:cursor-not-allowed';
  }

  return (
    <a
      href="#"
      className={className}
      onClick={(e) => {
        e.preventDefault();
        if (!row.getIsSelected()) {
          setDeletingTokens([row.original]);
          setShowDeleteTokenModal(true);
        }
      }}
    >
      Delete
      <span className="sr-only">, {row.original.name}</span>
    </a>
  );
}

function IndeterminateCheckbox({
  indeterminate,
  className = '',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = React.useRef<HTMLInputElement>(null!);

  useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate, rest.checked]);

  return (
    <input
      type="checkbox"
      ref={ref}
      className={
        className +
        ' text-purple-600 bg-gray-100 border-gray-300 h-4 w-4 cursor-pointer rounded focus:ring-purple-500'
      }
      {...rest}
    />
  );
}

type TokenTableProps = {
  tokens: IAuthToken[];
  setDeletingTokens: (tokens: IAuthToken[]) => void;
  setShowDeleteTokenModal: (show: boolean) => void;
};

export default function TokenTable(props: TokenTableProps) {
  const { tokens, setDeletingTokens, setShowDeleteTokenModal } = props;
  const searchThreshold = 10;
  const [sorting, setSorting] = useState<SortingState>([]);
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
  });

  const [filter, setFilter] = useState<string>('');
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const columnHelper = createColumnHelper<IAuthToken>();

  const deleteMultipleTokens = (table: Table<IAuthToken>) => {
    setDeletingTokens(table.getSelectedRowModel().rows.map((r) => r.original));
    setShowDeleteTokenModal(true);
  };

  useEffect(() => {
    setRowSelection({});
  }, [tokens]);

  const columns = [
    columnHelper.display({
      id: 'select',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler()
          }}
        />
      ),
      cell: ({ row }) => (
        <IndeterminateCheckbox
          {...{
            checked: row.getIsSelected(),
            disabled: !row.getCanSelect(),
            indeterminate: row.getIsSomeSelected(),
            onChange: row.getToggleSelectedHandler()
          }}
        />
      ),
      meta: {
        className: 'whitespace-nowrap py-4 pl-3 pr-4 text-left'
      }
    }),
    columnHelper.accessor('name', {
      header: ({ table }) => {
        if (table.getSelectedRowModel().rows.length > 0) {
          return (
            <Button
              onClick={(e) => {
                e.stopPropagation();
                deleteMultipleTokens(table);
              }}
              title="Delete Selected Token(s)"
            >
              Delete
            </Button>
          );
        }
        return 'Name';
      },
      cell: (info) => info.getValue(),
      meta: {
        className:
          'truncate whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-600'
      }
    }),
    columnHelper.accessor('description', {
      header: 'Description',
      cell: (info) => info.getValue(),
      meta: {
        className: 'truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500'
      }
    }),
    columnHelper.accessor('namespaceKey', {
      header: 'Namespace',
      cell: (info) => info.getValue(),
      meta: {
        className:
          'truncate whitespace-nowrap px-3 py-4 font-medium text-sm text-gray-500'
      }
    }),
    columnHelper.accessor(
      (row) => {
        if (!row.expiresAt) {
          return '';
        }
        return format(parseISO(row.expiresAt), 'MM/dd/yyyy');
      },
      {
        header: 'Expires',
        id: 'expiresAt',
        meta: {
          className:
            'truncate whitespace-nowrap px-3 py-4 font-semibold text-sm text-gray-500'
        }
      }
    ),
    columnHelper.accessor(
      (row) => format(parseISO(row.createdAt), 'MM/dd/yyyy'),
      {
        header: 'Created',
        id: 'createdAt',
        meta: {
          className: 'whitespace-nowrap px-3 py-4 text-sm text-gray-500'
        }
      }
    ),
    columnHelper.display({
      id: 'actions',
      cell: ({ row }) => (
        <TokenRowActions
          // eslint-disable-next-line react/prop-types
          row={row}
          setDeletingTokens={setDeletingTokens}
          setShowDeleteTokenModal={setShowDeleteTokenModal}
        />
      ),
      meta: {
        className:
          'whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6'
      }
    })
  ];

  const table = useReactTable({
    data: tokens,
    columns,
    state: {
      pagination,
      globalFilter: filter,
      sorting,
      rowSelection
    },
    globalFilterFn: 'includesString',
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onRowSelectionChange: setRowSelection,
    onPaginationChange: setPagination,
    getPaginationRowModel: getPaginationRowModel()
  });

  return (
    <>
      {(tokens.length >= searchThreshold || filter != '') && (
        <Searchbox className="mb-6" value={filter ?? ''} onChange={setFilter} />
      )}
      <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div className="relative overflow-hidden md:rounded-md">
            <table className="min-w-full table-fixed divide-y divide-gray-300">
              <thead className="bg-gray-50">
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
              <tbody className="bg-white divide-y divide-gray-200">
                {table.getRowModel().rows.map((row) => (
                  <tr key={row.id}>
                    {row.getVisibleCells().map((cell) => (
                      <td
                        key={cell.id}
                        className={cell.column.columnDef?.meta?.className}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
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
          </div>
        </div>
      </div>
    </>
  );
}
