import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import {
  CellContext,
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
import { format, parseISO } from 'date-fns';
import { useState } from 'react';
import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';
import { INamespace } from '~/types/Namespace';

type NamespaceEditActionProps = {
  cell: CellContext<INamespace, string>;
  setEditingNamespace: (token: INamespace) => void;
  setShowEditNamespaceModal: (show: boolean) => void;
};

function NamespaceEditAction(props: NamespaceEditActionProps) {
  const { cell, setEditingNamespace, setShowEditNamespaceModal } = props;

  return (
    <a
      href="#"
      className="text-violet-600 hover:text-violet-900"
      onClick={(e) => {
        e.preventDefault();
        setEditingNamespace(cell.row.original);
        setShowEditNamespaceModal(true);
      }}
    >
      {cell.getValue()}
    </a>
  );
}

type NamespaceDeleteActionProps = {
  row: Row<INamespace>;
  setDeletingNamespace: (token: INamespace) => void;
  setShowDeleteNamespaceModal: (show: boolean) => void;
};

function NamespaceDeleteAction(props: NamespaceDeleteActionProps) {
  const { row, setDeletingNamespace, setShowDeleteNamespaceModal } = props;
  return row.original.protected ? (
    <span
      title="Cannot deleting the default namespace"
      className="text-gray-400 hover:cursor-not-allowed"
    >
      Delete
      <span className="sr-only">, {row.original.name}</span>
    </span>
  ) : (
    <a
      href="#"
      className="text-violet-600 hover:text-violet-900"
      onClick={(e) => {
        e.preventDefault();
        setDeletingNamespace(row.original);
        setShowDeleteNamespaceModal(true);
      }}
    >
      Delete
      <span className="sr-only">, {row.original.name}</span>
    </a>
  );
}

type NamespaceTableProps = {
  namespaces: INamespace[];
  setEditingNamespace: (namespace: INamespace) => void;
  setShowEditNamespaceModal: (show: boolean) => void;
  setDeletingNamespace: (namespace: INamespace) => void;
  setShowDeleteNamespaceModal: (show: boolean) => void;
  readOnly?: boolean;
};

export default function NamespaceTable(props: NamespaceTableProps) {
  const {
    namespaces,
    setEditingNamespace,
    setShowEditNamespaceModal,
    setDeletingNamespace,
    setShowDeleteNamespaceModal,
    readOnly = false
  } = props;

  const searchThreshold = 10;

  const [sorting, setSorting] = useState<SortingState>([]);

  const [filter, setFilter] = useState<string>('');

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20
  });

  const columnHelper = createColumnHelper<INamespace>();

  const columns = [
    columnHelper.accessor('key', {
      header: 'Key',
      cell: (info) => info.getValue(),
      meta: {
        className:
          'truncate whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-900'
      }
    }),
    columnHelper.accessor('name', {
      header: 'Name',
      cell: (info) => {
        if (readOnly) {
          return info.getValue();
        }
        return (
          <NamespaceEditAction
            // eslint-disable-next-line react/prop-types
            cell={info}
            setEditingNamespace={setEditingNamespace}
            setShowEditNamespaceModal={setShowEditNamespaceModal}
          />
        );
      },
      meta: {
        className:
          'truncate whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-500'
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
        if (!row.updatedAt) {
          return '';
        }
        return format(parseISO(row.updatedAt), 'MM/dd/yyyy');
      },
      {
        header: 'Updated',
        id: 'updatedAt',
        meta: {
          className:
            'truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500'
        }
      }
    ),
    columnHelper.display({
      id: 'actions',
      cell: (props) => {
        if (!readOnly) {
          return (
            <NamespaceDeleteAction
              // eslint-disable-next-line react/prop-types
              row={props.row}
              setDeletingNamespace={setDeletingNamespace}
              setShowDeleteNamespaceModal={setShowDeleteNamespaceModal}
            />
          );
        }
      },
      meta: {
        className:
          'whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6'
      }
    })
  ];

  const table = useReactTable({
    data: namespaces,
    columns,
    state: {
      globalFilter: filter,
      sorting,
      pagination
    },
    globalFilterFn: 'includesString',
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onPaginationChange: setPagination,
    getPaginationRowModel: getPaginationRowModel()
  });

  return (
    <>
      {namespaces.length >= searchThreshold && (
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
