import { Button } from '~/components/Button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '~/components/Dialog';

import { useError } from '~/data/hooks/error';

type DeletePanelProps = {
  panelMessage: string | React.ReactNode;
  open: boolean;
  setOpen: (open: boolean) => void;
  handleDelete: (...args: string[]) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
  onError?: () => void;
};

export default function DeletePanel(props: DeletePanelProps) {
  const {
    open,
    setOpen,
    panelType,
    panelMessage,
    onSuccess,
    handleDelete,
    onError
  } = props;
  const { setError, clearError } = useError();

  const handleSubmit = () => {
    return handleDelete();
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete {panelType}</DialogTitle>
          <DialogDescription>
            This action is <span className="underline">destructive</span> and
            cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <div className="my-2">{panelMessage}</div>
        <DialogFooter>
          <DialogClose>Cancel</DialogClose>
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
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
