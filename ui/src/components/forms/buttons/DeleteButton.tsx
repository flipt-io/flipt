import { classNames } from '~/utils/helpers';

export type DeleteButtonProps = {
  onClick: () => void;
  disabled?: boolean;
  title?: string;
};

export function DeleteButton(props: DeleteButtonProps) {
  const { onClick, disabled, title } = props;
  return (
    <button
      type="button"
      className={classNames(
        disabled ? 'cursor-not-allowed' : 'cursor-hand',
        'text-red-400 border-red-200 mb-1 mt-5 inline-flex items-center justify-center rounded-md border px-4 py-2 text-sm font-medium enabled:hover:bg-red-50 focus:outline-none sm:mt-0'
      )}
      onClick={() => {
        !disabled && onClick && onClick();
      }}
      disabled={disabled}
      title={title}
    >
      Delete
    </button>
  );
}
