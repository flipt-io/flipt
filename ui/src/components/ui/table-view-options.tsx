import { DropdownMenuTrigger } from '@radix-ui/react-dropdown-menu';
import { Table } from '@tanstack/react-table';
import { ArrowDownNarrowWide, ArrowUpNarrowWide } from 'lucide-react';

import { Button } from '~/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent
} from '~/components/ui/dropdown-menu';

interface DataTableViewOptionsProps<TData> {
  table: Table<TData>;
}

export function DataTableViewOptions<TData>({
  table
}: DataTableViewOptionsProps<TData>) {
  const sorting = table.getState().sorting[0];

  const sortingColumn = table
    .getAllFlatColumns()
    .find((e) => e.id === sorting?.id);
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="ml-auto h-8 capitalize lg:flex"
        >
          {sorting?.desc ? <ArrowUpNarrowWide /> : <ArrowDownNarrowWide />}
          {sorting ? sortingColumn?.columnDef.header?.toString() : 'Sort'}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[150px]">
        {table
          .getAllColumns()
          .filter(
            (column) =>
              typeof column.accessorFn !== 'undefined' &&
              column.getCanSort() &&
              column.columnDef.header
          )
          .map((column) => {
            return (
              <DropdownMenuCheckboxItem
                key={column.id}
                className="capitalize"
                checked={column.getIsSorted() !== false}
                onCheckedChange={() => {
                  column.toggleSorting(column.getIsSorted() == 'asc', false);
                }}
              >
                {column.columnDef.header?.toString()}
              </DropdownMenuCheckboxItem>
            );
          })}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
