import { Transition } from '@headlessui/react';
import { XCircleIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { useError } from '~/data/hooks/error';

export default function ErrorNotification() {
  const { error, clearError } = useError();

  return (
    <Transition show={error !== null}>
      <div className="max-w-s fixed bottom-0 right-2 z-50 m-4 w-1/3">
        <div
          className="rounded-md border border-red-100 bg-red-50 p-4 shadow-xs"
          role="alert"
        >
          <div className="flex">
            <div className="shrink-0">
              <XCircleIcon
                className="h-5 w-5 text-red-400"
                aria-hidden="true"
              />
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              {error && (
                <div className="mt-2 text-sm text-red-700">{error}</div>
              )}
            </div>
            <div className="ml-auto pl-10">
              <div className="-mx-1 -my-1">
                <button
                  type="button"
                  onClick={() => {
                    clearError();
                  }}
                  className="inline-flex rounded-md bg-red-50 p-1.5 text-red-500 hover:bg-red-100 focus:outline-hidden focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-red-50"
                >
                  <span className="sr-only">Dismiss</span>
                  <XMarkIcon className="h-4 w-4" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  );
}
