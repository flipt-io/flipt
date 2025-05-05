import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/24/outline';
import {
  CellContext,
  PaginationState,
  Row,
  SortingState,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { XIcon } from 'lucide-react';
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
      className="text-violet-600 hover:text-violet-900 dark:text-violet-400 dark:hover:text-violet-300"
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
    <div className="flex justify-end">
      <span
        title="Cannot delete the default namespace"
        className="text-gray-400 dark:text-gray-500 hover:cursor-not-allowed"
      >
        <XIcon className="h-4 w-4" />
        <span className="sr-only">Cannot delete, {row.original.name}</span>
      </span>
    </div>
  ) : (
    <div className="flex justify-end">
      <button
        onClick={(e) => {
          e.stopPropagation();
          setDeletingNamespace(row.original);
          setShowDeleteNamespaceModal(true);
        }}
        className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400"
      >
        <XIcon className="h-4 w-4" />
        <span className="sr-only">Delete, {row.original.name}</span>
      </button>
    </div>
  );
}

type NamespaceTableProps = {
  namespaces: INamespace[];
  setEditingNamespace: (namespace: INamespace) => void;
  setShowEditNamespaceModal: (show: boolean) => void;
  setDeletingNamespace: (namespace: INamespace) => void;
  setShowDeleteNamespaceModal: (show: boolean) => void;
};

export default function NamespaceTable(props: NamespaceTableProps) {
  const {
    namespaces,
    setEditingNamespace,
    setShowEditNamespaceModal,
    setDeletingNamespace,
    setShowDeleteNamespaceModal
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
      cell: (info) => (
        <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
          {info.getValue()}
        </code>
      ),
      meta: {
        className:
          'whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-900 dark:text-gray-100'
      }
    }),
    columnHelper.accessor('name', {
      header: 'Name',
      cell: (info) => {
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
          'whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-500 dark:text-gray-400'
      }
    }),
    columnHelper.accessor('description', {
      header: 'Description',
      cell: (info) =>
        info.getValue() || (
          <span className="text-gray-400 dark:text-gray-500">—</span>
        ),
      meta: {
        className: 'px-3 py-4 text-sm text-gray-700 dark:text-gray-300'
      }
    }),
    columnHelper.display({
      id: 'actions',
      cell: (props) => {
        return (
          <NamespaceDeleteAction
            // eslint-disable-next-line react/prop-types
            row={props.row}
            setDeletingNamespace={setDeletingNamespace}
            setShowDeleteNamespaceModal={setShowDeleteNamespaceModal}
          />
        );
      },
      meta: {
        className: 'whitespace-nowrap px-3 py-4 text-sm font-medium'
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
      <div className="mt-4 overflow-x-auto">
        <div className="inline-block min-w-full py-2 align-middle">
          <div className="overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-800">
                {table.getHeaderGroups().map((headerGroup) => (
                  <tr key={headerGroup.id}>
                    {headerGroup.headers.map((header) =>
                      header.column.getCanSort() ? (
                        <th
                          key={header.id}
                          scope="col"
                          className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
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
                            <span className="ml-2 flex-none rounded text-gray-400 dark:text-gray-500 group-hover:visible group-focus:visible">
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
                          className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
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
              <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                {table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
                    onClick={() => {
                      setEditingNamespace(row.original);
                      setShowEditNamespaceModal(true);
                    }}
                  >
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
