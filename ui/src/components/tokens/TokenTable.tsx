import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  Table,
  SortingState,
  RowSelectionState,
  useReactTable
} from '@tanstack/react-table';
import { format, parseISO } from 'date-fns';
import React, { HTMLProps, useState } from 'react';
import Searchbox from '~/components/Searchbox';
import { IAuthToken } from '~/types/auth/Token';

type TokenRowActionsProps = {
  row: Row<IAuthToken>;
  setDeletingToken: (token: IAuthToken) => void;
  setShowDeleteTokenModal: (show: boolean) => void;
};

function TokenRowActions(props: TokenRowActionsProps) {
  const { row, setDeletingToken, setShowDeleteTokenModal } = props;

  let classes ="text-violet-600 hover:text-violet-900"
  if (row.getIsSelected()) {
    classes = "text-gray-400 hover:cursor-not-allowed"
  }

  return (
    <a
      href="#"
      className={classes}
      onClick={(e) => {
        e.preventDefault();
        if (!row.getIsSelected()){
          setDeletingToken(row.original);
          setShowDeleteTokenModal(true);
        }
      }}
    >
      Delete
      <span className="sr-only">, {row.original.name}</span>
    </a>
  );
}

type TokenTableProps = {
  tokens: IAuthToken[];
  setDeletingToken: (token: IAuthToken) => void;
  setShowDeleteTokenModal: (show: boolean) => void;
};


export default function TokenTable(props: TokenTableProps) {
  const { tokens, setDeletingToken, setShowDeleteTokenModal } = props;
  const searchThreshold = 10;
  const [sorting, setSorting] = useState<SortingState>([]);

  const [filter, setFilter] = useState<string>('');
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const columnHelper = createColumnHelper<IAuthToken>();

  const columns = [
     {
      id: 'select',
      header: ({ table }: Table) => <IndeterminateCheckbox
      {...{
        checked: table.getIsAllRowsSelected(),
        indeterminate: table.getIsSomeRowsSelected(),
        onChange: table.getToggleAllRowsSelectedHandler(),
      }}
       />,
      cell: ({row}: Row) => <IndeterminateCheckbox
      {...{
        checked: row.getIsSelected(),
        disabled: !row.getCanSelect(),
        indeterminate: row.getIsSomeSelected(),
        onChange: row.getToggleSelectedHandler(),
      }}
      />,
      meta: {
        className:
          'whitespace-nowrap py-4 pl-3 pr-4 text-left text-sm font-medium sm:pr-6'
      }
    },
    columnHelper.accessor('name', {
      header: 'Name',
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
      cell: ({row}: Row) => (
        <TokenRowActions
          // eslint-disable-next-line react/prop-types
          row={row}
          setDeletingToken={setDeletingToken}
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
  });

  return (
    <>
          <button
          onClick={() => console.info('selection', table.getSelectedRowModel().flatRows)}
        >
          Log `rowSelection` state
        </button>

      {tokens.length >= searchThreshold && (
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
          </div>
        </div>
      </div>
    </>
  );
}

function IndeterminateCheckbox({
  indeterminate,
  className = '',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = React.useRef<HTMLInputElement>(null!)

  React.useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate
    }
  }, [ref, indeterminate])

  return (
    <input
      type="checkbox"
      ref={ref}
      className={className + ' cursor-pointer'}
      {...rest}
    />
  )
}
