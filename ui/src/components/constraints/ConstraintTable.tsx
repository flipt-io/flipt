import {
  SortingState,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable
} from '@tanstack/react-table';
import {
  ArrowDownIcon,
  ArrowUpIcon,
  CalendarIcon,
  FilterIcon,
  HashIcon,
  IdCardIcon,
  PencilIcon,
  Text,
  ToggleLeftIcon,
  XIcon
} from 'lucide-react';
import { useState } from 'react';

import {
  ConstraintOperators,
  ConstraintType,
  IConstraint,
  NoValueOperators,
  constraintTypeToLabel
} from '~/types/Constraint';

import { useTimezone } from '~/data/hooks/timezone';

// Component for displaying array values in constraints
function ConstraintArrayValue({ value }: { value: string | undefined }) {
  try {
    const items: string[] | number[] = JSON.parse(value || '[]');
    return (
      <div className="max-w-[300px]">
        <div className="flex flex-wrap gap-1">
          {items.slice(0, 5).map((item, idx) => (
            <span
              key={idx}
              className="rounded-full bg-gray-100 dark:bg-gray-800 px-2 py-0.5 text-xs"
            >
              {String(item)}
            </span>
          ))}
          {items.length > 5 && (
            <span className="rounded-full bg-gray-100 dark:bg-gray-800 px-2 py-0.5 text-xs">
              +{items.length - 5} more
            </span>
          )}
        </div>
      </div>
    );
  } catch (err) {
    return <span className="text-red-500">Invalid array</span>;
  }
}

// Component for displaying constraint values based on type
function ConstraintValue({ constraint }: { constraint: IConstraint }) {
  const { inTimezone } = useTimezone();
  const isArrayValue = ['isoneof', 'isnotoneof'].includes(constraint.operator);

  if (
    constraint.type === ConstraintType.DATETIME &&
    constraint.value !== undefined
  ) {
    try {
      // Attempt to format the date - if it fails, show a fallback
      const formattedDate = inTimezone(constraint.value);
      return (
        <span className="text-sm text-gray-900 dark:text-white">
          {formattedDate}
        </span>
      );
    } catch (err) {
      // Show the raw value with an error indication
      return (
        <span className="text-sm text-red-600 dark:text-red-400">
          {constraint.value || '(invalid date)'}
        </span>
      );
    }
  }

  if (isArrayValue) {
    return <ConstraintArrayValue value={constraint.value} />;
  }

  if (constraint.type === ConstraintType.BOOLEAN) {
    const boolValue = constraint.value?.toLowerCase() === 'true';
    return (
      <span
        className={`text-sm ${
          boolValue
            ? 'text-emerald-600 dark:text-emerald-400'
            : 'text-red-600 dark:text-red-400'
        }`}
      >
        {boolValue ? 'TRUE' : 'FALSE'}
      </span>
    );
  }

  return (
    <span className="text-sm text-gray-900 dark:text-white break-words max-w-full">
      {constraint.value}
    </span>
  );
}

// Helper for type icon
function getTypeIcon(type: ConstraintType) {
  switch (type) {
    case ConstraintType.STRING:
      return <Text className="h-4 w-4" />;
    case ConstraintType.NUMBER:
      return <HashIcon className="h-4 w-4" />;
    case ConstraintType.BOOLEAN:
      return <ToggleLeftIcon className="h-4 w-4" />;
    case ConstraintType.DATETIME:
      return <CalendarIcon className="h-4 w-4" />;
    case ConstraintType.ENTITY_ID:
      return <IdCardIcon className="h-4 w-4" />;
    default:
      return <FilterIcon className="h-4 w-4" />;
  }
}

// Helper for type color
function getTypeColor(type: ConstraintType) {
  switch (type) {
    case ConstraintType.STRING:
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100';
    case ConstraintType.NUMBER:
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100';
    case ConstraintType.BOOLEAN:
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-100';
    case ConstraintType.DATETIME:
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-100';
    case ConstraintType.ENTITY_ID:
      return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-100';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100';
  }
}

type ConstraintTableProps = {
  constraints: IConstraint[];
  onEdit: (constraint: IConstraint, index: number) => void;
  onDelete: (constraint: IConstraint, index: number) => void;
};

export default function ConstraintTable({
  constraints,
  onEdit,
  onDelete
}: ConstraintTableProps) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'property', desc: false }
  ]);

  const columnHelper = createColumnHelper<IConstraint>();

  const columns = [
    columnHelper.accessor('type', {
      header: 'Type',
      cell: (info) => (
        <div className="flex items-center">
          <span
            className={`p-1 rounded-md mr-2 ${getTypeColor(info.getValue())}`}
          >
            {getTypeIcon(info.getValue())}
          </span>
          {constraintTypeToLabel(info.getValue())}
        </div>
      ),
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-sm font-medium'
      }
    }),
    columnHelper.accessor('property', {
      header: 'Property',
      cell: (info) => (
        <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
          {info.getValue()}
        </code>
      ),
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-sm'
      }
    }),
    columnHelper.accessor('operator', {
      header: 'Operator',
      cell: (info) => ConstraintOperators[info.getValue()] || info.getValue(),
      meta: {
        className: 'px-3 py-4 whitespace-nowrap text-sm'
      }
    }),
    columnHelper.accessor('value', {
      header: 'Value',
      cell: (info) => {
        const constraint = info.row.original;
        if (NoValueOperators.includes(constraint.operator)) {
          return <span className="text-gray-400 dark:text-gray-500">—</span>;
        }
        return <ConstraintValue constraint={constraint} />;
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
                // Find the original index in the constraints array
                const index = constraints.findIndex(
                  (c) =>
                    c.property === info.row.original.property &&
                    c.type === info.row.original.type &&
                    c.operator === info.row.original.operator &&
                    c.value === info.row.original.value
                );
                onDelete(info.row.original, index);
              }}
              className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              aria-label="Delete constraint"
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
    data: constraints,
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
              onClick={() => {
                // Find the original index in the constraints array
                const index = constraints.findIndex(
                  (c) =>
                    c.property === row.original.property &&
                    c.type === row.original.type &&
                    c.operator === row.original.operator &&
                    c.value === row.original.value
                );
                onEdit(row.original, index);
              }}
              tabIndex={0}
              role="button"
              aria-label={`Edit constraint for ${row.original.property}`}
              aria-pressed="false"
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  // Find the original index in the constraints array
                  const index = constraints.findIndex(
                    (c) =>
                      c.property === row.original.property &&
                      c.type === row.original.type &&
                      c.operator === row.original.operator &&
                      c.value === row.original.value
                  );
                  onEdit(row.original, index);
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
