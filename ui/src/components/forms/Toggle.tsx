import { Switch } from '@headlessui/react';
import { useField } from 'formik';
import { cls } from '~/utils/helpers';

type ToggleProps = {
  id: string;
  name: string;
  label: string;
  description?: string;
  checked: boolean;
  disabled?: boolean;
  onChange?: (e: any) => void;
};

export default function Toggle(props: ToggleProps) {
  const { id, label, description, checked, disabled = false, onChange } = props;
  const [field] = useField(props);

  return (
    <Switch.Group as="div" className="flex items-center justify-between">
      <span className="flex flex-grow flex-col">
        <Switch.Label
          as="span"
          className="text-sm font-medium text-gray-900"
          passive
        >
          {label}
        </Switch.Label>
        {description && (
          <Switch.Description as="span" className="text-sm text-gray-500">
            {description}
          </Switch.Description>
        )}
      </span>
      <Switch
        disabled={disabled}
        checked={checked}
        id={id}
        {...field}
        onChange={(e: boolean) => {
          onChange && onChange(e);
        }}
        className={cls(
          'relative inline-flex h-6 w-11 items-center rounded-full bg-violet-200 hover:cursor-pointer focus:ring-0',
          {
            'bg-green-400': checked,
            'hover:cursor-not-allowed': disabled
          }
        )}
      >
        <span className="sr-only">Enable</span>
        <span
          className={cls(
            'inline-block h-4 w-4 translate-x-1 transform rounded-full bg-background ring-0 transition',
            {
              'translate-x-6': checked
            }
          )}
        />
      </Switch>
    </Switch.Group>
  );
}
