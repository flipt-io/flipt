import { Skeleton } from '~/components/Skeleton';

const TableSkeleton = () => {
  return (
    <>
      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center justify-between">
          <Skeleton className="h-10 w-full max-w-60 text-xs lg:max-w-xs" />
          <Skeleton className="h-10 w-[75px] text-xs" />
        </div>
      </div>
      <Skeleton className="h-[96px] w-full mt-2" />
    </>
  );
};

TableSkeleton.displayName = 'TableSkeleton';

export { TableSkeleton };
