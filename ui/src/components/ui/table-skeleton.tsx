import { Skeleton } from '~/components/ui/skeleton';

const TableSkeleton = () => {
  return (
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between">
          <Skeleton className="h-8 w-full max-w-60 text-xs lg:max-w-md" />
          <Skeleton className="h-8 w-[75px] text-xs" />
        </div>
      </div>
      <Skeleton className="h-[96px] w-full" />
    </>
  );
};

TableSkeleton.displayName = 'TableSkeleton';

export { TableSkeleton };
