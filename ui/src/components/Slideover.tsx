import { Dialog, Transition } from '@headlessui/react';
import { forwardRef, Fragment } from 'react';

type SlideOverProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  children: React.ReactNode;
};

const SlideOver = forwardRef((props: SlideOverProps, ref: any) => {
  const { open, setOpen } = props;

  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog
        as="div"
        className="relative z-20"
        onClose={setOpen}
        initialFocus={ref}
      >
        <div className="fixed inset-0" />

        <div className="fixed inset-0 overflow-hidden">
          <div className="absolute inset-0 overflow-hidden">
            <div className="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10 sm:pl-16">
              <Transition.Child
                as={Fragment}
                enter="transform transition ease-in-out duration-200"
                enterFrom="translate-x-full"
                enterTo="translate-x-0"
                leave="transform transition ease-in-out duration-200"
                leaveFrom="translate-x-0"
                leaveTo="translate-x-full"
              >
                <Dialog.Panel className="pointer-events-auto w-screen max-w-xl">
                  {props.children}
                </Dialog.Panel>
              </Transition.Child>
            </div>
          </div>
        </div>
      </Dialog>
    </Transition.Root>
  );
});

SlideOver.displayName = 'SlideOver';
export default SlideOver;
