import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import * as Dialog from '@radix-ui/react-dialog';

import { Button } from '~/components/Button';

import { useError } from '~/data/hooks/error';

type DeletePanelProps = {
  panelMessage: string | React.ReactNode;
  setOpen: (open: boolean) => void;
  handleDelete: (...args: string[]) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
  onError?: () => void;
};

export default function DeletePanel(props: DeletePanelProps) {
  const { setOpen, panelType, panelMessage, onSuccess, handleDelete, onError } =
    props;
  const { setError, clearError } = useError();

  const handleSubmit = () => {
    return handleDelete();
  };

  return (
    <>
      <div className="sm:flex sm:items-start">
        <div className="mx-auto flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-red-100 dark:bg-red-900 sm:mx-0 sm:h-10 sm:w-10">
          <ExclamationTriangleIcon
            className="h-6 w-6 text-red-600 dark:text-red-300"
            aria-hidden="true"
          />
        </div>
        <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
          <Dialog.Title className="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
            Delete {panelType}
          </Dialog.Title>
          <div className="mt-2">
            <p className="text-sm text-gray-500 dark:text-gray-300">{panelMessage}</p>
          </div>
        </div>
      </div>
      <div className="mt-5 flex flex-row-reverse space-x-2 space-x-reverse sm:mt-4">
        <Button
          variant="destructive"
          type="button"
          onClick={() => {
            handleSubmit()
              ?.then(() => {
                clearError();
                if (onSuccess) {
                  onSuccess();
                }
              })
              .catch((err) => {
                setError(err);
                if (onError) {
                  onError();
                }
              })
              .finally(() => {
                setOpen(false);
              });
          }}
        >
          Delete
        </Button>
        <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
      </div>
    </>
  );
}
