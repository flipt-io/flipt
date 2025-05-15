import { useField } from 'formik';

import { Switch } from '~/components/Switch';

type ToggleProps = {
  id: string;
  name: string;
  label?: string;
  description?: string;
  checked: boolean;
  disabled?: boolean;
  onChange?: (e: any) => void;
};

export default function Toggle(props: ToggleProps) {
  const { id, label, description, checked, disabled = false, onChange } = props;
  const [field] = useField(props);

  return (
    <div className="flex items-center justify-between">
      <span className="flex grow flex-col">
        {label && (
          <span
            className="text-sm font-medium text-gray-900 dark:text-gray-100"
            id={'switch-label:' + id}
          >
            {label}
          </span>
        )}
        {description && (
          <span className="text-sm text-muted-foreground">{description}</span>
        )}
      </span>
      <Switch
        disabled={disabled}
        checked={checked}
        aria-labelledby={'switch-label:' + id}
        id={id}
        {...field}
        onCheckedChange={(e: boolean) => {
          onChange && onChange(e);
        }}
        className="data-[state=checked]:bg-success data-[state=unchecked]:bg-brand"
      />
    </div>
  );
}
