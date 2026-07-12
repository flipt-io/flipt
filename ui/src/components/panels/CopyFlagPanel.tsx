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
import { INamespace, INamespaceList } from '~/types/Namespace';
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
  const [remoteNamespaces, setRemoteNamespaces] = useState<INamespace[]>([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);
  const [remoteNamespacesLoaded, setRemoteNamespacesLoaded] = useState(false);
  const [remoteNamespacesLoadError, setRemoteNamespacesLoadError] =
    useState<string>();
  const [remoteNamespacesRetry, setRemoteNamespacesRetry] = useState(0);

  useEffect(() => {
    if (open) {
      return;
    }

    setSelectedEnvironment(undefined);
    setSelectedNamespace(undefined);
    setRemoteNamespaces((current) => (current.length > 0 ? [] : current));
    setNamespacesLoading(false);
    setRemoteNamespacesLoaded(false);
    setRemoteNamespacesLoadError(undefined);
    setRemoteNamespacesRetry(0);
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

  useEffect(() => {
    if (!open || !selectedEnvironment?.key) {
      return;
    }

    let active = true;

    if (selectedEnvironment.key === environment.key) {
      setRemoteNamespaces((current) => (current.length > 0 ? [] : current));
      setNamespacesLoading((current) => (current ? false : current));
      setRemoteNamespacesLoaded(false);
      setRemoteNamespacesLoadError(undefined);
      return;
    }

    setNamespacesLoading(true);
    setRemoteNamespacesLoaded(false);
    setRemoteNamespacesLoadError(undefined);
    request('GET', `${apiURL}/${selectedEnvironment.key}/namespaces`)
      .then((response) => {
        if (!active) {
          return;
        }

        const data = response as INamespaceList;

        setRemoteNamespaces(data.items);
        setRemoteNamespacesLoaded(true);
      })
      .catch((err) => {
        if (!active) {
          return;
        }

        setRemoteNamespaces([]);
        setRemoteNamespacesLoaded(false);
        setRemoteNamespacesLoadError(
          err instanceof Error
            ? err.message
            : 'Unable to load namespaces for the selected environment.'
        );
      })
      .finally(() => {
        if (active) {
          setNamespacesLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [environment.key, open, remoteNamespacesRetry, selectedEnvironment?.key]);

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
              key={selectedEnvironment?.key || 'copy-namespace'}
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
                  onClick={() => setRemoteNamespacesRetry((retry) => retry + 1)}
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
