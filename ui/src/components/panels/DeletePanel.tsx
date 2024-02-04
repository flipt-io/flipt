import { Dialog } from '@headlessui/react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import Button from '~/components/forms/buttons/Button';
import { useError } from '~/data/hooks/error';

type DeletePanelProps = {
  panelMessage: string | React.ReactNode;
  setOpen: (open: boolean) => void;
  handleDelete: (...args: string[]) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
};

export default function DeletePanel(props: DeletePanelProps) {
  const { setOpen, panelType, panelMessage, onSuccess, handleDelete } = props;
  const { setError, clearError } = useError();

  const handleSubmit = () => {
    return handleDelete();
  };

  return (
    <>
      <div className="sm:flex sm:items-start">
        <div className="bg-red-100 mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full sm:mx-0 sm:h-10 sm:w-10">
          <ExclamationTriangleIcon
            className="text-red-600 h-6 w-6"
            aria-hidden="true"
          />
        </div>
        <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
          <Dialog.Title
            as="h3"
            className="text-gray-900 text-lg font-medium leading-6"
          >
            Delete {panelType}
          </Dialog.Title>
          <div className="mt-2">
            <p className="text-gray-500 text-sm">{panelMessage}</p>
          </div>
        </div>
      </div>
      <div className="mt-5 flex flex-row-reverse space-x-2 space-x-reverse sm:mt-4">
        <Button
          variant="primary"
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
              })
              .finally(() => {
                setOpen(false);
              });
          }}
        >
          Delete
        </Button>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
      </div>
    </>
  );
}
