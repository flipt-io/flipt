import { Transition } from '@headlessui/react';
import {
  BellAlertIcon,
  SparklesIcon,
  XMarkIcon
} from '@heroicons/react/24/outline';
import { Fragment, useState } from 'react';
import { useSessionStorage } from '~/data/hooks/storage';
import { IInfo } from '~/types/Meta';

type NotificationProps = {
  show: boolean;
  setShow: (show: boolean) => void;
  markSeen: () => void;
  info: IInfo;
};

export function Notification(props: NotificationProps) {
  const { info, show, setShow, markSeen } = props;

  return (
    <>
      <div
        aria-live="assertive"
        className="pointer-events-none fixed inset-0 z-10 flex items-end px-4 py-6 sm:items-start sm:p-4"
      >
        <div className="flex w-full flex-col items-center space-y-4 sm:items-end">
          <Transition
            show={show}
            as={Fragment}
            enter="transform ease-out duration-300 transition"
            enterFrom="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
            enterTo="translate-y-0 opacity-100 sm:translate-x-0"
            leave="transition ease-in duration-100"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="bg-white pointer-events-auto w-full max-w-sm overflow-hidden rounded-lg shadow-lg ring-1 ring-black ring-opacity-5">
              <div className="p-4">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <SparklesIcon
                      className="text-gray-400 h-6 w-6"
                      aria-hidden="true"
                    />
                  </div>
                  <div className="ml-3 w-0 flex-1 pt-0.5">
                    <p className="text-gray-900 text-sm font-medium">
                      Update Available
                    </p>
                    <p className="text-gray-500 mt-1 text-sm">
                      A new version of Flipt is available!
                    </p>
                    <div className="mt-3 flex space-x-7 hover:cursor-pointer">
                      <a
                        href={info.latestVersionURL}
                        target="_blank"
                        rel="noreferrer"
                        className="bg-white text-violet-600 rounded-md text-sm font-medium hover:text-violet-500 focus:outline-none"
                      >
                        Check It Out
                      </a>
                      <a
                        className="bg-white text-gray-700 rounded-md text-sm font-medium hover:text-gray-500 focus:outline-none"
                        onClick={(e) => {
                          e.preventDefault();
                          setShow(false);
                          markSeen();
                        }}
                      >
                        Dismiss
                      </a>
                    </div>
                  </div>
                  <div className="ml-4 flex flex-shrink-0">
                    <button
                      type="button"
                      className="bg-white text-gray-400 inline-flex rounded-md hover:text-gray-500 focus:outline-none"
                      onClick={() => {
                        setShow(false);
                        markSeen();
                      }}
                    >
                      <span className="sr-only">Close</span>
                      <XMarkIcon className="h-5 w-5" aria-hidden="true" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </>
  );
}

type NotificationsProps = {
  info: IInfo;
};

export default function Notifications(props: NotificationsProps) {
  const { info } = props;
  const [show, setShow] = useState(false);

  const [newNotifications, setNewNotification] = useSessionStorage(
    'new_notifications',
    info.updateAvailable
  );

  return (
    <>
      <Notification
        info={info}
        show={show}
        setShow={setShow}
        markSeen={() => setNewNotification(false)}
      />

      <button
        type="button"
        className="without-ring text-violet-100 relative rounded-full"
      >
        {newNotifications && (
          <span className="absolute right-0 top-0 flex h-2 w-2">
            <span className="bg-violet-100 relative inline-flex h-2 w-2 rounded-full"></span>
          </span>
        )}
        <span className="sr-only">View notifications</span>

        <BellAlertIcon
          className="nightwind-prevent text-gray-300 h-5 w-5 hover:text-gray-100 dark:text-gray-400 dark:hover:text-gray-200"
          aria-hidden="true"
          onClick={() => {
            setShow(true);
          }}
        />
      </button>
    </>
  );
}
