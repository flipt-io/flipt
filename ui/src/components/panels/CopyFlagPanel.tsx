import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

import {
  selectCurrentEnvironment,
  selectEnvironments,
  useListEnvironmentsQuery
} from '~/app/environments/environmentsApi';
import {
  selectCurrentNamespace,
  selectNamespaces,
  useListNamespacesQuery
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

import { useError } from '~/data/hooks/error';
import { getErrorMessage } from '~/utils/helpers';

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
    displayValue: namespace.name || namespace.key
  };
}

export default function CopyFlagPanel(props: CopyFlagPanelProps) {
  const { open, setOpen, panelMessage, onSuccess, handleCopy } = props;

  const { setError, clearError } = useError();

  const environment = useSelector(selectCurrentEnvironment);
  const environmentsFromStore = useSelector(selectEnvironments);
  const namespace = useSelector(selectCurrentNamespace);
  const namespacesFromStore = useSelector(selectNamespaces);
  const { data: environmentsData } = useListEnvironmentsQuery();
  const { data: namespacesData } = useListNamespacesQuery(
    {
      environmentKey: environment.key
    },
    { skip: !environment.key }
  );

  const environments = useMemo(
    () =>
      (environmentsData?.environments ?? environmentsFromStore).filter(
        (candidate) => !candidate.configuration?.base
      ),
    [environmentsData?.environments, environmentsFromStore]
  );
  const namespaces = namespacesData?.items ?? namespacesFromStore;

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
  const shouldLoadRemoteNamespaces =
    open &&
    !!selectedEnvironment?.key &&
    selectedEnvironment.key !== environment.key;
  const {
    data: remoteNamespacesData,
    error: remoteNamespacesError,
    isFetching: remoteNamespacesLoading,
    isError: remoteNamespacesIsError,
    refetch: refetchRemoteNamespaces
  } = useListNamespacesQuery(
    {
      environmentKey: selectedEnvironment?.key || ''
    },
    { skip: !shouldLoadRemoteNamespaces }
  );
  const remoteNamespaces = useMemo(
    () => remoteNamespacesData?.items ?? [],
    [remoteNamespacesData?.items]
  );
  const namespacesLoading =
    shouldLoadRemoteNamespaces && remoteNamespacesLoading;
  const remoteNamespacesLoaded =
    shouldLoadRemoteNamespaces && remoteNamespacesData !== undefined;
  const remoteNamespacesLoadError =
    shouldLoadRemoteNamespaces && remoteNamespacesIsError
      ? getErrorMessage(remoteNamespacesError)
      : undefined;

  useEffect(() => {
    if (open) {
      return;
    }

    setSelectedEnvironment(undefined);
    setSelectedNamespace(undefined);
  }, [open]);

  useEffect(() => {
    if (!open || environmentOptions.length === 0) {
      return;
    }

    const initialEnvironmentKey = hasAlternativeNamespace
      ? environment.key
      : environmentOptions.find(
          (candidate) => candidate.key !== environment.key
        )?.key || environment.key;

    setSelectedEnvironment((current) => {
      if (
        current &&
        environmentOptions.some((candidate) => candidate.key === current.key)
      ) {
        return current;
      }

      return (
        environmentOptions.find(
          (candidate) => candidate.key === initialEnvironmentKey
        ) || environmentOptions[0]
      );
    });
  }, [open, environment.key, environmentOptions, hasAlternativeNamespace]);

  const namespaceOptions = useMemo(() => {
    const availableNamespaces =
      selectedEnvironment?.key === environment.key
        ? namespaces
        : remoteNamespaces;

    return availableNamespaces
      .filter((candidate) => {
        if (selectedEnvironment?.key !== environment.key) {
          return true;
        }

        return candidate.key !== namespace.key;
      })
      .map(toSelectableNamespace);
  }, [
    environment.key,
    namespace.key,
    namespaces,
    remoteNamespaces,
    selectedEnvironment?.key
  ]);

  useEffect(() => {
    if (!open) {
      return;
    }

    if (namespaceOptions.length === 0) {
      if (selectedNamespace) {
        setSelectedNamespace(undefined);
      }
      return;
    }

    const next =
      (selectedNamespace
        ? namespaceOptions.find(
            (candidate) => candidate.key === selectedNamespace.key
          )
        : undefined) || namespaceOptions[0];

    if (selectedNamespace?.key !== next.key) {
      setSelectedNamespace(next);
    }
  }, [open, namespaceOptions, selectedNamespace]);

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
    !selectedEnvironment ||
    !selectedNamespace ||
    namespacesLoading ||
    !!remoteNamespacesLoadError;
  const showEmptyNamespaces =
    !namespacesLoading &&
    !remoteNamespacesLoadError &&
    namespaceOptions.length === 0 &&
    (selectedEnvironment?.key === environment.key || remoteNamespacesLoaded);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Copy Flag</DialogTitle>
          <DialogDescription>{panelMessage}</DialogDescription>
        </DialogHeader>
        <div className="my-2 space-y-4">
          <div className="space-y-2">
            <div id="copyToEnvironment-label" className="text-sm font-medium">
              Environment
            </div>
            <Listbox<SelectableEnvironment>
              id="copyToEnvironment"
              name="environmentKey"
              ariaLabelledBy="copyToEnvironment-label"
              values={environmentOptions}
              selected={selectedEnvironment}
              setSelected={setSelectedEnvironment}
            />
          </div>
          <div className="space-y-2">
            <div id="copyToNamespace-label" className="text-sm font-medium">
              Namespace
            </div>
            <Listbox<SelectableNamespace>
              id="copyToNamespace"
              name="namespaceKey"
              ariaLabelledBy="copyToNamespace-label"
              values={namespaceOptions}
              selected={selectedNamespace}
              setSelected={setSelectedNamespace}
              disabled={
                namespacesLoading ||
                !!remoteNamespacesLoadError ||
                namespaceOptions.length === 0
              }
            />
            {remoteNamespacesLoadError && (
              <div className="space-y-2 rounded-md border border-destructive/30 bg-destructive/10 p-3">
                <p className="text-sm text-destructive">
                  {remoteNamespacesLoadError}
                </p>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => refetchRemoteNamespaces()}
                >
                  Retry
                </Button>
              </div>
            )}
            {showEmptyNamespaces && (
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
