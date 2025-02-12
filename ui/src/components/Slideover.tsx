import * as Dialog from '@radix-ui/react-dialog';
import { forwardRef } from 'react';

type SlideOverProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  children: React.ReactNode;
};

const SlideOver = forwardRef((props: SlideOverProps, ref: any) => {
  const { open, setOpen } = props;

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-20" />
        <Dialog.Content
          ref={ref}
          className="fixed inset-y-0 right-0 z-20 flex max-w-full pl-10 sm:pl-16"
        >
          <div className="pointer-events-none fixed inset-y-0 right-0 flex max-w-full">
            <div
              className="pointer-events-auto w-screen max-w-xl transform transition-transform duration-200 ease-in-out"
              data-state={open ? 'open' : 'closed'}
              style={{
                transform: open ? 'translateX(0)' : 'translateX(100%)'
              }}
            >
              {props.children}
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
});

SlideOver.displayName = 'SlideOver';
export default SlideOver;
