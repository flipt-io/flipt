import { FolderOpenIcon } from 'lucide-react';
import { useSelector } from 'react-redux';
import { useLocation, useNavigate } from 'react-router';

import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/Select';

import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';

import { useAppDispatch } from '~/data/hooks/store';
import { addNamespaceToPath, cls } from '~/utils/helpers';

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

  const setCurrentNamespace = (namespace: INamespace) => {
    dispatch(currentNamespaceChanged(namespace));
    const newPath = addNamespaceToPath(location.pathname, namespace.key);
    navigate(newPath);
  };

  const changeNamespace = (ns: string) => {
    const value = namespaces?.find((el) => el.key == ns) as INamespace;
    value && setCurrentNamespace(value);
  };

  return (
    <div className="flex items-center gap-1" data-testid="namespace-listbox">
      <FolderOpenIcon
        className="ml-2 h-6 w-6 shrink-0 text-white md:text-gray-500"
        aria-hidden="true"
      />
      <Select
        disabled={disabled || namespaces.length < 2}
        defaultValue={namespace.key}
        onValueChange={changeNamespace}
      >
        <SelectTrigger
          className={cls(
            'bg-gray-200 dark:bg-gray-100 border-0 focus:outline-none focus-visible:ring-0',
            className
          )}
          aria-label={namespace.key}
        >
          <SelectValue placeholder="Select a namespace" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {namespaces?.map((v) => (
              <SelectItem key={v.key} value={v.key}>
                {v.name}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  );
}
