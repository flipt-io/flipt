import { useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

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
import Listbox from '~/components/forms/Listbox';

import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';

import { useError } from '~/data/hooks/error';

export type SelectableNamespace = Pick<INamespace, 'key' | 'name'> &
  ISelectable;

type CopyToNamespacePanelProps = {
  panelMessage: string | React.ReactNode;
  open: boolean;
  setOpen: (open: boolean) => void;
  handleCopy: (namespaceKey: string) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
};

export default function CopyToNamespacePanel(props: CopyToNamespacePanelProps) {
  const { open, setOpen, panelType, panelMessage, onSuccess, handleCopy } =
    props;

  const { setError, clearError } = useError();

  const namespace = useSelector(selectCurrentNamespace);

  // get all namespaces except the current one
  const namespaces = useSelector(selectNamespaces).filter(
    (n) => n.key !== namespace.key
  );

  const [selectedNamespace, setSelectedNamespace] =
    useState<SelectableNamespace>({
      ...namespaces[0],
      displayValue: namespaces[0]?.name
    });

  const handleSubmit = () => {
    return handleCopy(selectedNamespace.key);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Copy {panelType}</DialogTitle>
          <DialogDescription>{panelMessage}</DialogDescription>
        </DialogHeader>
        <Listbox<SelectableNamespace>
          id="copyToNamespace"
          name="namespaceKey"
          values={namespaces.map((n) => ({
            ...n,
            displayValue: n.name
          }))}
          selected={{
            ...selectedNamespace,
            displayValue: selectedNamespace?.name || ''
          }}
          setSelected={setSelectedNamespace}
          className="my-2"
        />
        <DialogFooter>
          <DialogClose>Cancel</DialogClose>
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
                  setOpen(false);
                })
                .catch((err) => {
                  setError(err);
                });
            }}
          >
            Copy
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
