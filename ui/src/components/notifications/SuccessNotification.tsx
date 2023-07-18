import { Transition } from '@headlessui/react';
import { CheckCircleIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { useEffect } from 'react';
import { useSuccess } from '~/data/hooks/success';

export default function SuccessNotification() {
  const { success, clearSuccess } = useSuccess();

  // Close the notification after 3 seconds
  useEffect(() => {
    if (success !== null) {
      setTimeout(() => clearSuccess(), 3000);
    }
  }, [success, clearSuccess]);

  return (
    <Transition show={success !== null}>
      <div className="max-w-s fixed bottom-0 right-2 z-10 m-4">
        <div className="bg-green-50 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <CheckCircleIcon
                className="text-green-400 h-5 w-5"
                aria-hidden="true"
              />
            </div>
            <div className="ml-3">
              <p className="text-green-800 text-sm font-medium">{success}</p>
            </div>
            <div className="ml-auto pl-3">
              <div className="-mx-1.5 -my-1.5">
                <button
                  type="button"
                  onClick={() => clearSuccess()}
                  className="text-green-500 bg-green-50 inline-flex rounded-md p-1.5 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-green-600 focus:ring-offset-2 focus:ring-offset-green-50"
                >
                  <span className="sr-only">Dismiss</span>
                  <XMarkIcon className="h-5 w-5" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  );
}
