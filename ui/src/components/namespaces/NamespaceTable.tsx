import {
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
import { ArrowDownIcon, ArrowUpIcon, PencilIcon, XIcon } from 'lucide-react';
import { useMemo, useState } from 'react';

import Pagination from '~/components/Pagination';
import Searchbox from '~/components/Searchbox';

import { INamespace } from '~/types/Namespace';

type NamespaceDeleteActionProps = {
  row: Row<INamespace>;
  setEditingNamespace: (namespace: INamespace) => void;
  setShowEditNamespaceModal: (show: boolean) => void;
  setDeletingNamespace: (namespace: INamespace) => void;
  setShowDeleteNamespaceModal: (show: boolean) => void;
};

function NamespaceDeleteAction(props: NamespaceDeleteActionProps) {
  const {
    row,
    setEditingNamespace,
    setShowEditNamespaceModal,
    setDeletingNamespace,
    setShowDeleteNamespaceModal
  } = props;

  return (
    <div className="flex items-center justify-end gap-2">
      <div className="hidden group-hover/row:block">
        <span
          className="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded flex items-center gap-1 cursor-pointer"
          onClick={(e) => {
            e.stopPropagation();
            setEditingNamespace(row.original);
            setShowEditNamespaceModal(true);
          }}
        >
          <PencilIcon className="h-3.5 w-3.5" />
          Edit
        </span>
      </div>
      {row.original.protected ? (
        <div className="relative group/delete inline-block">
          <span className="text-gray-400 dark:text-gray-500 hover:cursor-not-allowed p-1 rounded-full">
            <XIcon className="h-4 w-4" />
          </span>
        </div>
      ) : (
        <div className="relative group/delete inline-block">
          <button
            onClick={(e) => {
              e.stopPropagation();
              setDeletingNamespace(row.original);
              setShowDeleteNamespaceModal(true);
            }}
            className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            aria-label={`Delete namespace ${row.original.name}`}
          >
            <XIcon className="h-4 w-4" />
          </button>
        </div>
      )}
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

  const sortedNamespaces = useMemo(() => {
    return [...namespaces].sort((a, b) => {
      if (a.key.toLowerCase() === 'default') return -1;
      if (b.key.toLowerCase() === 'default') return 1;
      return a.name.localeCompare(b.name);
    });
  }, [namespaces]);

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
      cell: (info) => info.getValue(),
      meta: {
        className:
          'whitespace-nowrap px-3 py-4 text-sm font-medium text-gray-500 dark:text-gray-400'
      }
    }),
    columnHelper.accessor('description', {
      header: 'Description',
      cell: (info) =>
        info.getValue() ? (
          <div className="truncate max-w-full">{info.getValue()}</div>
        ) : (
          <span className="text-gray-400 dark:text-gray-500">—</span>
        ),
      meta: {
        className:
          'px-3 py-4 text-sm text-gray-700 dark:text-gray-300 max-w-[300px]'
      }
    }),
    columnHelper.display({
      id: 'actions',
      header: '',
      cell: (props) => {
        return (
          <NamespaceDeleteAction
            // eslint-disable-next-line react/prop-types
            row={props.row}
            setEditingNamespace={setEditingNamespace}
            setShowEditNamespaceModal={setShowEditNamespaceModal}
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
    data: sortedNamespaces,
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
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 table-fixed w-full">
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
                              {header.column.getIsSorted() === 'asc' ? (
                                <ArrowUpIcon
                                  className="h-4 w-4 text-gray-500 dark:text-gray-400"
                                  aria-hidden="true"
                                />
                              ) : header.column.getIsSorted() === 'desc' ? (
                                <ArrowDownIcon
                                  className="h-4 w-4 text-gray-500 dark:text-gray-400"
                                  aria-hidden="true"
                                />
                              ) : (
                                <div className="h-4 w-4 opacity-0 group-hover:opacity-40">
                                  <ArrowUpIcon className="h-4 w-4" />
                                </div>
                              )}
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
                    className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer group/row relative focus-within:bg-gray-50 dark:focus-within:bg-gray-800 focus-within:outline-none focus-within:ring-2 focus-within:ring-violet-500 dark:focus-within:ring-violet-400"
                    onClick={() => {
                      setEditingNamespace(row.original);
                      setShowEditNamespaceModal(true);
                    }}
                    tabIndex={0}
                    role="button"
                    aria-label={`Edit namespace ${row.original.name}`}
                    aria-pressed="false"
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setEditingNamespace(row.original);
                        setShowEditNamespaceModal(true);
                      }
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
