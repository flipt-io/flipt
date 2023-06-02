import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  SortingState,
  useReactTable
} from '@tanstack/react-table';
import { format, parseISO } from 'date-fns';
import { useState } from 'react';
import Searchbox from '~/components/Searchbox';
import { IAuthToken } from '~/types/auth/Token';

type TokenRowActionsProps = {
  row: Row<IAuthToken>;
  setDeletingToken: (token: IAuthToken) => void;
  setShowDeleteTokenModal: (show: boolean) => void;
};

function TokenRowActions(props: TokenRowActionsProps) {
  const { row, setDeletingToken, setShowDeleteTokenModal } = props;
  return (
    <a
      href="#"
      className="text-violet-600 hover:text-violet-900"
      onClick={(e) => {
        e.preventDefault();
        setDeletingToken(row.original);
        setShowDeleteTokenModal(true);
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

  const columnHelper = createColumnHelper<IAuthToken>();

  const columns = [
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
            'truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500'
        }
      }
    ),
    columnHelper.display({
      id: 'actions',
      cell: (props) => (
        <TokenRowActions
          // eslint-disable-next-line react/prop-types
          row={props.row}
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
      sorting
    },
    globalFilterFn: 'includesString',
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel()
  });

  return (
    <>
      {tokens.length >= searchThreshold && (
        <Searchbox className="mb-6" value={filter ?? ''} onChange={setFilter} />
      )}
      <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div className="relative overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-md">
            <table className="min-w-full table-fixed divide-y divide-gray-300">
              <thead className="bg-gray-50">
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
              <tbody className="divide-y divide-gray-200 bg-white">
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
