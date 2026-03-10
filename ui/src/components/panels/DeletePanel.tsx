import { AlertTriangle } from 'lucide-react';
import { Button } from '~/components/Button';
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
        <div className="mx-auto flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
          <AlertTriangle
            className="text-destructive h-6 w-6"
            aria-hidden="true"
          />
        </div>
        <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
          <h3 className="text-secondary-foreground text-lg leading-6 font-medium">
            Delete {panelType}
          </h3>
          <div className="mt-2">
            <p className="text-muted-foreground text-sm">{panelMessage}</p>
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
