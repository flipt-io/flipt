import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectAllEnvironments,
  selectCurrentEnvironment
} from '~/app/environments/environmentsApi';
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
import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';

import { apiURL, request } from '~/data/api';
import { useError } from '~/data/hooks/error';

type SelectableEnvironment = Pick<IEnvironment, 'key' | 'name'> & ISelectable;
type SelectableNamespace = Pick<INamespace, 'key' | 'name'> & ISelectable;

type CopyFlagPanelProps = {
  panelMessage: string | React.ReactNode;
  open: boolean;
  setOpen: (open: boolean) => void;
  handleCopy: (target: {
    environmentKey: string;
    namespaceKey: string;
  }) => Promise<any>;
  onSuccess?: () => void;
};

function toSelectableEnvironment(
  environment: IEnvironment
): SelectableEnvironment {
  return {
    key: environment.key,
    name: environment.name,
    displayValue: environment.name || environment.key
  };
}

function toSelectableNamespace(namespace: INamespace): SelectableNamespace {
  return {
    key: namespace.key,
    name: namespace.name,
    displayValue: namespace.name
  };
}

export default function CopyFlagPanel(props: CopyFlagPanelProps) {
  const { open, setOpen, panelMessage, onSuccess, handleCopy } = props;

  const { setError, clearError } = useError();

  const environment = useSelector(selectCurrentEnvironment);
  const environments = useSelector(selectAllEnvironments);
  const namespace = useSelector(selectCurrentNamespace);
  const namespaces = useSelector(selectNamespaces);

  const environmentOptions = useMemo(
    () => environments.map(toSelectableEnvironment),
    [environments]
  );

  const hasAlternativeNamespace = useMemo(
    () => namespaces.some((candidate) => candidate.key !== namespace.key),
    [namespaces, namespace.key]
  );

  const [selectedEnvironment, setSelectedEnvironment] =
    useState<SelectableEnvironment>();
  const [selectedNamespace, setSelectedNamespace] =
    useState<SelectableNamespace>();
  const [namespaceOptions, setNamespaceOptions] = useState<
    SelectableNamespace[]
  >([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);

  useEffect(() => {
    if (!open || environmentOptions.length === 0) {
      return;
    }

    const initialEnvironmentKey = hasAlternativeNamespace
      ? environment.key
      : environmentOptions.find(
          (candidate) => candidate.key !== environment.key
        )?.key || environment.key;

    const initialEnvironment =
      environmentOptions.find(
        (candidate) => candidate.key === initialEnvironmentKey
      ) || environmentOptions[0];

    setSelectedEnvironment(initialEnvironment);
  }, [open, environment.key, environmentOptions, hasAlternativeNamespace]);

  useEffect(() => {
    if (!open || !selectedEnvironment?.key) {
      return;
    }

    let active = true;

    setNamespacesLoading(true);

    request('GET', `${apiURL}/${selectedEnvironment.key}/namespaces`)
      .then((response) => {
        if (!active) {
          return;
        }

        const data = response as { items?: INamespace[] };

        setNamespaceOptions(
          (data.items ?? [])
            .filter((candidate) => {
              if (selectedEnvironment.key !== environment.key) {
                return true;
              }

              return candidate.key !== namespace.key;
            })
            .map(toSelectableNamespace)
        );
      })
      .catch((err) => {
        if (!active) {
          return;
        }

        setNamespaceOptions([]);
        setError(err);
      })
      .finally(() => {
        if (active) {
          setNamespacesLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [environment.key, namespace.key, open, selectedEnvironment, setError]);

  useEffect(() => {
    if (!open) {
      return;
    }

    if (namespaceOptions.length === 0) {
      setSelectedNamespace(undefined);
      return;
    }

    setSelectedNamespace((current) => {
      if (!current) {
        return namespaceOptions[0];
      }

      return (
        namespaceOptions.find((candidate) => candidate.key === current.key) ||
        namespaceOptions[0]
      );
    });
  }, [open, namespaceOptions]);

  const handleSubmit = () => {
    if (!selectedEnvironment || !selectedNamespace) {
      return Promise.resolve();
    }

    return handleCopy({
      environmentKey: selectedEnvironment.key,
      namespaceKey: selectedNamespace.key
    });
  };

  const isCopyDisabled =
    !selectedEnvironment || !selectedNamespace || namespacesLoading;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Copy Flag</DialogTitle>
          <DialogDescription>{panelMessage}</DialogDescription>
        </DialogHeader>
        <div className="my-2 space-y-4">
          <div className="space-y-2">
            <div className="text-sm font-medium">Environment</div>
            <Listbox<SelectableEnvironment>
              id="copyToEnvironment"
              name="environmentKey"
              values={environmentOptions}
              selected={selectedEnvironment}
              setSelected={setSelectedEnvironment}
            />
          </div>
          <div className="space-y-2">
            <div className="text-sm font-medium">Namespace</div>
            <Listbox<SelectableNamespace>
              key={selectedEnvironment?.key || 'copy-namespace'}
              id="copyToNamespace"
              name="namespaceKey"
              values={namespaceOptions}
              selected={selectedNamespace}
              setSelected={setSelectedNamespace}
              disabled={namespacesLoading || namespaceOptions.length === 0}
            />
            {namespaceOptions.length === 0 && (
              <p className="text-sm text-muted-foreground">
                No namespaces are available in the selected environment.
              </p>
            )}
          </div>
        </div>
        <DialogFooter>
          <DialogClose>Cancel</DialogClose>
          <Button
            variant="primary"
            type="button"
            disabled={isCopyDisabled}
            onClick={() => {
              handleSubmit()
                .then(() => {
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
