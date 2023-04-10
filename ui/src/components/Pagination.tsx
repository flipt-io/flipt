import {
  ArrowLongLeftIcon,
  ArrowLongRightIcon
} from '@heroicons/react/20/solid';
import { usePagination } from '~/data/hooks/pagination';
import { classNames } from '~/utils/helpers';

type PageProps = {
  page: number | string;
  currentPage: number;
  onPageChange: (page: number) => void;
};

function Page(props: PageProps) {
  const { page, currentPage, onPageChange } = props;
  // we are using '...' (string) to represent page links that should not be rendered
  if (typeof page === 'string') {
    return (
      <span className="border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-700">
        &#8230; {/* ellipsis */}
      </span>
    );
  }

  return (
    <a
      href="#"
      className={classNames(
        page === currentPage
          ? 'border-violet-500 text-violet-600'
          : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
        'inline-flex items-center border-t-2 px-4 pt-4 text-sm font-medium'
      )}
      aria-current={page === currentPage ? 'page' : undefined}
      onClick={(e) => {
        e.preventDefault();
        onPageChange(page);
      }}
    >
      {page}
    </a>
  );
}

export type PaginationProps = {
  className?: string;
  currentPage: number;
  totalCount: number;
  pageSize: number;
  onPageChange: (page: number) => void;
};

export default function Pagination(props: PaginationProps) {
  const {
    className = '',
    currentPage,
    totalCount,
    pageSize,
    onPageChange
  } = props;

  const paginationRange = usePagination({
    currentPage,
    totalCount,
    siblingCount: 2,
    pageSize
  });

  const onNextPage = () => {
    onPageChange(currentPage + 1);
  };

  const onPreviousPage = () => {
    onPageChange(currentPage - 1);
  };

  const lastPage = paginationRange[paginationRange.length - 1];

  return (
    <nav
      className={`${className} flex items-center justify-between border-t border-gray-200 px-4 sm:px-0`}
    >
      <div className="flex w-0 flex-1">
        {currentPage > 1 && (
          <a
            href="#"
            className="inline-flex items-center border-t-2 border-transparent pr-1 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700"
            onClick={(e) => {
              e.preventDefault();
              onPreviousPage();
            }}
          >
            <ArrowLongLeftIcon
              className="mr-3 h-5 w-5 text-gray-400"
              aria-hidden="true"
            />
            Previous
          </a>
        )}
      </div>
      <div className="hidden md:-mt-px md:flex">
        {paginationRange.map((page, i) => (
          <Page
            key={i}
            page={page}
            currentPage={currentPage}
            onPageChange={onPageChange}
          />
        ))}
      </div>
      <div className="-mt-px flex w-0 flex-1 justify-end">
        {currentPage < lastPage && (
          <a
            href="#"
            className="inline-flex items-center border-t-2 border-transparent pl-1 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700"
            onClick={(e) => {
              e.preventDefault();
              onNextPage();
            }}
          >
            Next
            <ArrowLongRightIcon
              className="ml-3 h-5 w-5 text-gray-400"
              aria-hidden="true"
            />
          </a>
        )}
      </div>
    </nav>
  );
}
