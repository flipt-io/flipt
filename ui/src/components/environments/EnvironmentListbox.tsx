import { useSelector } from 'react-redux';

import {
  currentEnvironmentChanged,
  selectCurrentEnvironment,
  selectEnvironments
} from '~/app/environments/environmentsApi';

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/Select';

import { IEnvironment } from '~/types/Environment';
import { ISelectable } from '~/types/Selectable';

import { useAppDispatch } from '~/data/hooks/store';
import { cls } from '~/utils/helpers';

export type SelectableEnvironment = Pick<IEnvironment, 'key'> & ISelectable;

type EnvironmentListboxProps = {
  className?: string;
};

export default function EnvironmentListbox(props: EnvironmentListboxProps) {
  const { className } = props;
  const environment = useSelector(selectCurrentEnvironment);
  const dispatch = useAppDispatch();
  const environments = useSelector(selectEnvironments);

  // Sort environments to show default first
  const sortedEnvironments = [...environments].sort((a, b) => {
    if (a.default) return -1;
    if (b.default) return 1;
    return a.key.localeCompare(b.key);
  });

  const changeEnvironment = (key: string) => {
    const env = environments?.find((el) => el.key == key) as IEnvironment;
    if (env) {
      dispatch(currentEnvironmentChanged(env));
    }
  };

  return (
    <Select
      disabled={environments.length <= 1}
      defaultValue={environment.key}
      onValueChange={changeEnvironment}
    >
      <SelectTrigger
        className={cls(
          'bg-black text-white uppercase border-0 focus:outline-none w-auth min-w-24 focus-visible:ring-0',
          className
        )}
        data-testid="environment-listbox"
      >
        <SelectValue placeholder="Select a namespace" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {sortedEnvironments?.map((v) => (
            <SelectItem
              key={v.key}
              value={v.key}
              className="uppercase min-w-[160px] overflow-auto "
            >
              {v.key || 'Unknown Environment'}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
