import { Transition } from '@headlessui/react';
import { XCircleIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { useError } from '~/data/hooks/error';

export default function ErrorNotification() {
  const { error, clearError } = useError();

  return (
    <Transition show={error !== null}>
      <div className="max-w-s fixed bottom-0 right-2 z-50 m-4 w-1/3">
        <div
          className="bg-red-50 border-red-100 rounded-md border p-4 shadow"
          role="alert"
        >
          <div className="flex">
            <div className="flex-shrink-0">
              <XCircleIcon
                className="text-red-400 h-5 w-5"
                aria-hidden="true"
              />
            </div>
            <div className="ml-3">
              <h3 className="text-red-800 text-sm font-medium">Error</h3>
              {error && (
                <div className="text-red-700 mt-2 text-sm">{error}</div>
              )}
            </div>
            <div className="ml-auto pl-10">
              <div className="-mx-1 -my-1">
                <button
                  type="button"
                  onClick={() => {
                    clearError();
                  }}
                  className="text-red-500 bg-red-50 inline-flex rounded-md p-1.5 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-green-50"
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
