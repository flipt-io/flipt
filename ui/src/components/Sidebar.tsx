import { Dialog, Transition } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Fragment } from 'react';
import { Link } from 'react-router-dom';
import logoLight from '~/assets/logo-light.png';
import Nav from './Nav';
import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';

type SidebarProps = {
  sidebarOpen: boolean;
  setSidebarOpen: (sidebarOpen: boolean) => void;
};

export default function Sidebar(props: SidebarProps) {
  const { sidebarOpen, setSidebarOpen } = props;

  const { ui } = useSelector(selectConfig);
  const topbarStyle = { backgroundColor: ui.topbar?.color };
  return (
    <>
      <Transition.Root show={sidebarOpen} as={Fragment}>
        <Dialog
          as="div"
          className="relative z-40 md:hidden"
          onClose={setSidebarOpen}
        >
          <Transition.Child
            as={Fragment}
            enter="transition-opacity ease-linear duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="transition-opacity ease-linear duration-300"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="bg-gray-300 fixed inset-0 bg-opacity-75" />
          </Transition.Child>

          <div className="fixed inset-0 z-40 flex">
            <Transition.Child
              as={Fragment}
              enter="transition ease-in-out duration-300 transform"
              enterFrom="-translate-x-full"
              enterTo="translate-x-0"
              leave="transition ease-in-out duration-300 transform"
              leaveFrom="translate-x-0"
              leaveTo="-translate-x-full"
            >
              <Dialog.Panel className="relative flex w-full max-w-xs flex-1 flex-col bg-gray-950 pb-4 pt-5">
                <Transition.Child
                  as={Fragment}
                  enter="ease-in-out duration-300"
                  enterFrom="opacity-0"
                  enterTo="opacity-100"
                  leave="ease-in-out duration-300"
                  leaveFrom="opacity-100"
                  leaveTo="opacity-0"
                >
                  <div className="absolute right-0 top-0 -mr-12 pt-2">
                    <button
                      type="button"
                      className="ml-1 flex h-10 w-10 items-center justify-center rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                      onClick={() => setSidebarOpen(false)}
                    >
                      <span className="sr-only">Close sidebar</span>
                      <XMarkIcon
                        className="nightwind-prevent text-white h-6 w-6"
                        aria-hidden="true"
                      />
                    </button>
                  </div>
                </Transition.Child>
                <div className="mt-2 flex flex-shrink-0 items-center px-4">
                  <img
                    src={logoLight}
                    alt="logo"
                    width={549}
                    height={191}
                    className="h-10 w-auto"
                  />
                </div>
                <div className="mt-5 h-0 flex-1 overflow-y-auto">
                  <Nav
                    sidebarOpen={sidebarOpen}
                    setSidebarOpen={setSidebarOpen}
                  />
                </div>
              </Dialog.Panel>
            </Transition.Child>
            <div className="w-14 flex-shrink-0" aria-hidden="true">
              {/* Dummy element to force sidebar to shrink to fit close icon */}
            </div>
          </div>
        </Dialog>
      </Transition.Root>

      {/* Static sidebar for desktop */}
      <div className="hidden md:fixed md:inset-y-0 md:flex md:w-64 md:flex-col">
        <div className="bg-gray-200 flex min-h-0 flex-1 flex-col">
          <div
            className="relative flex h-16 flex-shrink-0 items-center bg-gray-950 px-4 pb-1 pt-2 dark:border-b dark:border-b-white/20"
            style={topbarStyle}
          >
            <Link to="/">
              <img
                src={logoLight}
                alt="logo"
                width={549}
                height={191}
                className="h-10 w-auto"
              />
            </Link>
          </div>
          <div className="flex flex-1 flex-col overflow-y-auto">
            <Nav className="py-4" />
          </div>
        </div>
      </div>
    </>
  );
}
