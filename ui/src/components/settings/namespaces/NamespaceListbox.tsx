import { useLocation, useNavigate } from 'react-router-dom';
import Listbox, { ISelectable } from '~/components/forms/Listbox';
import useNamespace from '~/data/hooks/namespace';
import { INamespace } from '~/types/Namespace';
import { addNamespaceToPath } from '~/utils/helpers';

type SelectableNamespace = Pick<INamespace, 'key' | 'name'> & ISelectable;

type NamespaceLisboxProps = {
  disabled: boolean;
  namespaces: INamespace[];
  className?: string;
};

export default function NamespaceListbox(props: NamespaceLisboxProps) {
  const { disabled, namespaces, className } = props;

  const { currentNamespace } = useNamespace();
  const location = useLocation();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: SelectableNamespace) => {
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
        ...currentNamespace,
        displayValue: currentNamespace.name || ''
      }}
      setSelected={setCurrentNamespace}
    />
  );
}
