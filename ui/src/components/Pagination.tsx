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
      <span className="border-t-2 border-transparent px-4 pt-4 text-sm font-medium text-gray-700 dark:text-gray-400">
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
          'border-violet-500 text-violet-600 dark:text-violet-400':
            isCurrentPage,
          'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300':
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
      className={`${className} flex items-center justify-between border-t border-gray-200 dark:border-gray-700 px-4 sm:px-0`}
    >
      <div className="flex w-0 flex-1">
        {currentPage > 1 && (
          <a
            href="#"
            className="inline-flex items-center border-t-2 border-transparent pr-1 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300"
            onClick={(e) => {
              e.preventDefault();
              onPreviousPage();
            }}
          >
            <ArrowLongLeftIcon
              className="mr-3 h-5 w-5 text-gray-400 dark:text-gray-500"
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
          className="mt-4 inline-flex items-center border-transparent bg-gray-50 dark:bg-gray-800 text-sm font-medium text-gray-500 dark:text-gray-400 focus:border-transparent focus:ring-0"
          value={pageSize}
          onChange={(e) => {
            onPageChange(currentPage, Number(e.target.value));
          }}
        >
          {pageSizes.map((pageSize) => (
            <option
              key={pageSize}
              value={pageSize}
              className="inline-flex items-center border-transparent text-sm font-medium text-gray-500 dark:text-gray-400 focus:border-transparent focus:ring-0"
            >
              Show {pageSize}
            </option>
          ))}
        </select>
      </div>
      <div className="-mt-px flex w-0 flex-1 justify-end">
        {currentPage < (lastPage as number) && (
          <a
            href="#"
            className="inline-flex items-center border-t-2 border-transparent pl-1 pt-4 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300"
            onClick={(e) => {
              e.preventDefault();
              onNextPage();
            }}
          >
            Next
            <ArrowLongRightIcon
              className="ml-3 h-5 w-5 text-gray-400 dark:text-gray-500"
              aria-hidden="true"
            />
          </a>
        )}
      </div>
    </nav>
  );
}
