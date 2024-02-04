import { Dialog } from '@headlessui/react';
import { DocumentDuplicateIcon } from '@heroicons/react/24/outline';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import {
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import Button from '~/components/forms/buttons/Button';
import Listbox from '~/components/forms/Listbox';
import { SelectableNamespace } from '~/components/namespaces/NamespaceListbox';
import { useError } from '~/data/hooks/error';

type CopyToNamespacePanelProps = {
  panelMessage: string | React.ReactNode;
  setOpen: (open: boolean) => void;
  handleCopy: (namespaceKey: string) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
};

export default function CopyToNamespacePanel(props: CopyToNamespacePanelProps) {
  const { setOpen, panelType, panelMessage, onSuccess, handleCopy } = props;

  const { setError, clearError } = useError();

  const namespace = useSelector(selectCurrentNamespace);

  // get all namespaces except the current one
  const namespaces = useSelector(selectNamespaces).filter(
    (n) => n.key !== namespace.key
  );

  const [selectedNamespace, setSelectedNamespace] =
    useState<SelectableNamespace>({
      ...namespaces[0],
      displayValue: namespaces[0].name
    });

  const handleSubmit = () => {
    return handleCopy(selectedNamespace.key);
  };

  return (
    <>
      <div className="sm:flex sm:items-start">
        <div className="bg-violet-100 mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full sm:mx-0 sm:h-10 sm:w-10">
          <DocumentDuplicateIcon
            className="text-violet-500 h-6 w-6"
            aria-hidden="true"
          />
        </div>
        <div className="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
          <Dialog.Title
            as="h3"
            className="text-gray-900 text-lg font-medium leading-6"
          >
            Copy {panelType}
          </Dialog.Title>
          <div className="mt-2">
            <p className="text-gray-500 text-sm">{panelMessage}</p>
          </div>
          <div className="mt-4">
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
            />
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
                setOpen(false);
              })
              .catch((err) => {
                setError(err);
              });
          }}
        >
          Copy
        </Button>
        <Button onClick={() => setOpen(false)}>Cancel</Button>
      </div>
    </>
  );
}
