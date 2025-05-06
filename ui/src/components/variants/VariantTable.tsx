import {
  SortingState,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import { ArrowDownIcon, ArrowUpIcon, PencilIcon, XIcon } from 'lucide-react';
import { useState } from 'react';

import { IVariant } from '~/types/Variant';

type VariantTableProps = {
  variants: IVariant[];
  onEdit: (variant: IVariant) => void;
  onDelete: (variant: IVariant) => void;
};

export default function VariantTable({
  variants,
  onEdit,
  onDelete
}: VariantTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'key', desc: false }
  ]);

  const columnHelper = createColumnHelper<IVariant>();

  const columns = [
    columnHelper.accessor('key', {
      header: 'Key',
      cell: (info) => (
        <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
          {info.getValue()}
        </code>
      ),
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-sm'
      }
    }),
    columnHelper.accessor('name', {
      header: 'Name',
      cell: (info) => (
        <div className="truncate max-w-[200px]">
          {info.getValue() || (
            <span className="text-gray-400 dark:text-gray-500">—</span>
          )}
        </div>
      ),
      meta: {
        className:
          'px-3 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100'
      }
    }),
    columnHelper.accessor('attachment', {
      header: 'Attachment',
      cell: (info) => {
        const hasAttachment =
          info.getValue() && Object.keys(info.getValue() || {}).length > 0;

        return hasAttachment ? (
          <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-500 dark:text-gray-400 text-xs">
            {Object.keys(info.getValue() || {}).length} fields
          </span>
        ) : (
          <span className="text-gray-400 dark:text-gray-500">—</span>
        );
      },
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-sm'
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
      cell: (info) => (
        <div className="flex items-center justify-end gap-2">
          <div className="hidden group-hover/row:block">
            <span className="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded flex items-center gap-1">
              <PencilIcon className="h-3.5 w-3.5" />
              Edit
            </span>
          </div>
          <div className="relative group/delete inline-block">
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onDelete(info.row.original);
              }}
              className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              aria-label={`Delete variant ${info.row.original.key}`}
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
      ),
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-right text-sm font-medium'
      }
    })
  ];

  const table = useReactTable({
    data: variants,
    columns,
    state: {
      sorting
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel()
  });

  return (
    <div className="mt-4 overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700 table-fixed w-full">
        <thead className="bg-gray-50 dark:bg-gray-800">
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th
                  key={header.id}
                  scope="col"
                  className={`px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider ${
                    header.column.getCanSort()
                      ? 'cursor-pointer select-none'
                      : ''
                  }`}
                  onClick={header.column.getToggleSortingHandler()}
                >
                  {header.isPlaceholder ? null : (
                    <div className="flex items-center">
                      {flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                      {header.column.getCanSort() && (
                        <div className="ml-1">
                          {header.column.getIsSorted() === 'asc' ? (
                            <ArrowUpIcon className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                          ) : header.column.getIsSorted() === 'desc' ? (
                            <ArrowDownIcon className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                          ) : (
                            <div className="h-4 w-4 opacity-0 group-hover:opacity-40">
                              <ArrowUpIcon className="h-4 w-4" />
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {table.getRowModel().rows.map((row) => (
            <tr
              key={row.id}
              className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer group/row relative focus-within:bg-gray-50 dark:focus-within:bg-gray-800 focus-within:outline-none focus-within:ring-2 focus-within:ring-violet-500 dark:focus-within:ring-violet-400"
              onClick={() => onEdit(row.original)}
              tabIndex={0}
              role="button"
              aria-label={`Edit variant ${row.original.key}`}
              aria-pressed="false"
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onEdit(row.original);
                }
              }}
            >
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
    </div>
  );
}
