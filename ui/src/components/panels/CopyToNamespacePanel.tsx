import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentEnvironment,
  selectEnvironments
} from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';
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

import { IEnvironment } from '~/types/Environment';
import { Product } from '~/types/Meta';
import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';

import { useError } from '~/data/hooks/error';

export type SelectableNamespace = Pick<INamespace, 'key' | 'name'> &
  ISelectable;

type SelectableEnvironment = Pick<IEnvironment, 'key' | 'name'> & ISelectable;

const conflictStrategies = [
  { id: 'FAIL', label: 'Fail', description: 'Abort if the resource exists' },
  {
    id: 'OVERWRITE',
    label: 'Overwrite',
    description: 'Replace the existing resource'
  },
  {
    id: 'SKIP',
    label: 'Skip',
    description: 'Keep the existing resource'
  }
];

type CopyToNamespacePanelProps = {
  panelMessage: string | React.ReactNode;
  open: boolean;
  setOpen: (open: boolean) => void;
  handleCopy: (
    namespaceKey: string,
    environmentKey?: string,
    onConflict?: string
  ) => Promise<any>;
  panelType: string;
  onSuccess?: () => void;
};

export default function CopyToNamespacePanel(props: CopyToNamespacePanelProps) {
  const { open, setOpen, panelType, panelMessage, onSuccess, handleCopy } =
    props;

  const { setError, clearError } = useError();

  const info = useSelector(selectInfo);
  const isPro = info.product === Product.PRO;

  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const environments = useSelector(selectEnvironments);
  const namespace = useSelector(selectCurrentNamespace);
  const allNamespaces = useSelector(selectNamespaces);

  const [selectedEnvironment, setSelectedEnvironment] =
    useState<SelectableEnvironment>({
      ...currentEnvironment,
      displayValue: currentEnvironment?.name || currentEnvironment?.key || ''
    });

  const isCrossEnvironment = selectedEnvironment.key !== currentEnvironment.key;

  // Filter namespaces: exclude current namespace only when copying within the same environment
  const namespaces = useMemo(
    () =>
      isCrossEnvironment
        ? allNamespaces
        : allNamespaces.filter((n) => n.key !== namespace.key),
    [allNamespaces, namespace.key, isCrossEnvironment]
  );

  const [selectedNamespace, setSelectedNamespace] =
    useState<SelectableNamespace>({
      ...namespaces[0],
      displayValue: namespaces[0]?.name
    });

  const [onConflict, setOnConflict] = useState('FAIL');

  const showEnvironmentSelector = isPro && environments.length > 1;

  const handleSubmit = () => {
    const targetEnvKey = isCrossEnvironment
      ? selectedEnvironment.key
      : undefined;
    const conflictStrategy = isPro ? onConflict : undefined;
    return handleCopy(selectedNamespace.key, targetEnvKey, conflictStrategy);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Copy {panelType}</DialogTitle>
          <DialogDescription>{panelMessage}</DialogDescription>
        </DialogHeader>
        {showEnvironmentSelector && (
          <div>
            <label
              htmlFor="copyToEnvironment"
              className="block text-sm font-medium text-secondary-foreground mb-1"
            >
              Environment
            </label>
            <Listbox<SelectableEnvironment>
              id="copyToEnvironment"
              name="environmentKey"
              values={environments.map((e) => ({
                ...e,
                displayValue: e.name || e.key
              }))}
              selected={{
                ...selectedEnvironment,
                displayValue:
                  selectedEnvironment?.name || selectedEnvironment?.key || ''
              }}
              setSelected={(env) => {
                setSelectedEnvironment(env);
                // Reset namespace selection when environment changes
                const newIsCross = env.key !== currentEnvironment.key;
                const filteredNs = newIsCross
                  ? allNamespaces
                  : allNamespaces.filter((n) => n.key !== namespace.key);
                if (filteredNs.length > 0) {
                  setSelectedNamespace({
                    ...filteredNs[0],
                    displayValue: filteredNs[0].name
                  });
                }
              }}
              className="my-1"
            />
          </div>
        )}
        <div>
          <label
            htmlFor="copyToNamespace"
            className="block text-sm font-medium text-secondary-foreground mb-1"
          >
            Namespace
          </label>
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
            className="my-1"
          />
        </div>
        {isPro && (
          <div>
            <label className="block text-sm font-medium text-secondary-foreground mb-2">
              On Conflict
            </label>
            <div className="space-y-2">
              {conflictStrategies.map((strategy) => (
                <label
                  key={strategy.id}
                  className="flex items-start gap-2 cursor-pointer"
                >
                  <input
                    type="radio"
                    name="onConflict"
                    value={strategy.id}
                    checked={onConflict === strategy.id}
                    onChange={() => setOnConflict(strategy.id)}
                    className="mt-0.5 h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                  />
                  <div className="text-sm">
                    <span className="font-medium text-foreground">
                      {strategy.label}
                    </span>
                    <span className="text-muted-foreground ml-1">
                      - {strategy.description}
                    </span>
                  </div>
                </label>
              ))}
            </div>
          </div>
        )}
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
