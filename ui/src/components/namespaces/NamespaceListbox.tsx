import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  currentNamespace,
  currentNamespaceChanged,
  fetchNamespaces,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import Listbox, { ISelectable } from '~/components/forms/Listbox';
import { useError } from '~/data/hooks/error';
import { useAppDispatch, useAppSelector } from '~/data/hooks/store';
import { INamespace } from '~/types/Namespace';
import { addNamespaceToPath } from '~/utils/helpers';

type SelectableNamespace = Pick<INamespace, 'key' | 'name'> & ISelectable;

type NamespaceLisboxProps = {
  disabled: boolean;
  className?: string;
};

export default function NamespaceListbox(props: NamespaceLisboxProps) {
  const { disabled, className } = props;

  const { setError, clearError } = useError();

  const namespace = useSelector(currentNamespace);

  const dispatch = useAppDispatch();
  const namespaces = useSelector(selectNamespaces);
  const namespacesStatus = useAppSelector((state) => {
    return state.namespaces.status;
  });
  const error = useAppSelector((state) => {
    return state.namespaces.error;
  });

  useEffect(() => {
    if (namespacesStatus === 'idle') {
      dispatch(fetchNamespaces());
    }
  }, [namespacesStatus, dispatch]);

  useEffect(() => {
    if (error) {
      setError(Error(error));
      return;
    }
    clearError();
  }, [clearError, error, setError]);

  const location = useLocation();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: SelectableNamespace) => {
    dispatch(currentNamespaceChanged(namespace));
    // navigate to the current location.path with the new namespace prependend
    // e.g. /namespaces/default/segments -> /namespaces/namespaceKey/segments
    const newPath = addNamespaceToPath(location.pathname, namespace.key);
    navigate(newPath);
  };

  if (namespacesStatus === 'succeeded') {
    return (
      <Listbox<SelectableNamespace>
        disabled={disabled}
        id="namespaceKey"
        name="namespaceKey"
        className={className}
        values={namespaces.map((n) => ({
          ...n,
          displayValue: n.name
        }))}
        selected={{
          ...namespace,
          displayValue: namespace?.name || ''
        }}
        setSelected={setCurrentNamespace}
      />
    );
  }
}
