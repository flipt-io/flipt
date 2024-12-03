import { Skeleton } from '~/components/ui/skeleton';

const TableSkeleton = () => {
  return (
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between space-x-2 py-4">
          <Skeleton className="h-8 w-[150px] text-xs lg:w-[250px]" />
          <Skeleton className="h-8 w-[75px] text-xs" />
        </div>
      </div>
      <Skeleton className="h-[96px] w-full" />
    </>
  );
};

TableSkeleton.displayName = 'TableSkeleton';

export { TableSkeleton };
