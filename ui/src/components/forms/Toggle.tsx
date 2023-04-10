import { Switch } from '@headlessui/react';
import { useField } from 'formik';

type ToggleProps = {
  id: string;
  name: string;
  label: string;
  description?: string;
  enabled: boolean;
  handleChange?: (e: boolean) => void;
};

export default function Toggle(props: ToggleProps) {
  const { id, label, description, enabled, handleChange } = props;
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
        checked={enabled}
        id={id}
        {...field}
        onChange={(e: boolean) => {
          handleChange && handleChange(e);
        }}
        className={`${
          enabled ? 'bg-green-400' : 'bg-violet-200'
        } relative inline-flex h-6 w-11 items-center rounded-full`}
      >
        <span className="sr-only">Enable</span>
        <span
          className={`${
            enabled ? 'translate-x-6' : 'translate-x-1'
          } inline-block h-4 w-4 transform rounded-full bg-white transition`}
        />
      </Switch>
    </Switch.Group>
  );
}
