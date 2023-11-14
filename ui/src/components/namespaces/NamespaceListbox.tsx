import { useSelector } from 'react-redux';
import { useLocation, useNavigate } from 'react-router-dom';
import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesSlice';
import Listbox from '~/components/forms/Listbox';
import { useAppDispatch } from '~/data/hooks/store';
import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';
import { addNamespaceToPath } from '~/utils/helpers';

export type SelectableNamespace = Pick<INamespace, 'key' | 'name'> &
  ISelectable;

type NamespaceListboxProps = {
  disabled: boolean;
  className?: string;
};

export default function NamespaceListbox(props: NamespaceListboxProps) {
  const { disabled, className } = props;

  const namespace = useSelector(selectCurrentNamespace);

  const dispatch = useAppDispatch();
  const namespaces = useSelector(selectNamespaces);

  const location = useLocation();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: SelectableNamespace) => {
    dispatch(currentNamespaceChanged(namespace));
    // navigate to the current location.path with the new namespace prependend
    // e.g. /namespaces/default/segments -> /namespaces/namespaceKey/segments
    const newPath = addNamespaceToPath(location.pathname, namespace.key);
    navigate(newPath);
  };

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
