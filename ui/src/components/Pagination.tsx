import {
  ArrowLongLeftIcon,
  ArrowLongRightIcon
} from '@heroicons/react/20/solid';
import { useMemo } from 'react';
import { usePagination } from '~/data/hooks/pagination';
import { cls } from '~/utils/helpers';

type PageProps = {
  page: number | string;
  currentPage: number;
  onPageChange: (page: number) => void;
};

const pageSizes = [20, 50, 100];

function Page(props: PageProps) {
  const { page, currentPage, onPageChange } = props;
  const isCurrentPage = useMemo(
    () => page === currentPage,
    [page, currentPage]
  );

  // we are using '...' (string) to represent page links that should not be rendered
  if (typeof page === 'string') {
    return (
      <span className="text-gray-700 border-t-2 border-transparent px-4 pt-4 text-sm font-medium">
        &#8230; {/* ellipsis */}
      </span>
    );
  }

  return (
    <a
      href="#"
      className={cls(
        'inline-flex items-center border-t-2 px-4 pt-4 text-sm font-medium',
        {
          'text-violet-600 border-violet-500': isCurrentPage,
          'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300':
            !isCurrentPage
        }
      )}
      aria-current={isCurrentPage ? 'page' : undefined}
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
  onPageChange: (page: number, size: number) => void;
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
    onPageChange(currentPage + 1, pageSize);
  };

  const onPreviousPage = () => {
    onPageChange(currentPage - 1, pageSize);
  };

  const lastPage = paginationRange[paginationRange.length - 1];

  if (totalCount <= pageSizes[0]) {
    return null;
  }

  return (
    <nav
      className={`${className} border-gray-200 flex items-center justify-between border-t px-4 sm:px-0`}
    >
      <div className="flex w-0 flex-1">
        {currentPage > 1 && (
          <a
            href="#"
            className="text-gray-500 inline-flex items-center border-t-2 border-transparent pr-1 pt-4 text-sm font-medium hover:text-gray-700 hover:border-gray-300"
            onClick={(e) => {
              e.preventDefault();
              onPreviousPage();
            }}
          >
            <ArrowLongLeftIcon
              className="text-gray-400 mr-3 h-5 w-5"
              aria-hidden="true"
            />
            Previous
          </a>
        )}
      </div>
      {paginationRange.length > 1 && (
        <div className="hidden md:-mt-px md:flex">
          {paginationRange.map((page, i) => (
            <Page
              key={i}
              page={page}
              currentPage={currentPage}
              onPageChange={(page) => onPageChange(page, pageSize)}
            />
          ))}
        </div>
      )}
      <div className="hidden md:-mt-px md:flex">
        <select
          className="text-gray-500 bg-gray-50 mt-4 inline-flex items-center border-transparent text-sm font-medium focus:border-transparent focus:ring-0"
          value={pageSize}
          onChange={(e) => {
            onPageChange(currentPage, Number(e.target.value));
          }}
        >
          {pageSizes.map((pageSize) => (
            <option
              key={pageSize}
              value={pageSize}
              className="text-gray-500 inline-flex items-center border-transparent text-sm font-medium focus:border-transparent focus:ring-0"
            >
              Show {pageSize}
            </option>
          ))}
        </select>
      </div>
      <div className="-mt-px flex w-0 flex-1 justify-end">
        {currentPage < lastPage && (
          <a
            href="#"
            className="text-gray-500 inline-flex items-center border-t-2 border-transparent pl-1 pt-4 text-sm font-medium hover:text-gray-700 hover:border-gray-300"
            onClick={(e) => {
              e.preventDefault();
              onNextPage();
            }}
          >
            Next
            <ArrowLongRightIcon
              className="text-gray-400 ml-3 h-5 w-5"
              aria-hidden="true"
            />
          </a>
        )}
      </div>
    </nav>
  );
}
