import { cls } from '~/utils/helpers';

export type ButtonProps = {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  type?: 'button' | 'submit' | 'reset';
  variant?: 'primary' | 'secondary' | 'soft';
  className?: string;
  title?: string;
  disabled?: boolean;
};

export default function Button(props: ButtonProps) {
  const {
    className,
    onClick,
    children,
    type = 'button',
    variant = 'secondary',
    title,
    disabled = false
  } = props;

  return (
    <button
      type={type}
      onClick={(e) => {
        !disabled && onClick && onClick(e);
      }}
      className={cls(
        'cursor-hand mb-1 inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium focus:outline-none focus:ring-1 focus:ring-offset-1',
        className,
        {
          'cursor-not-allowed': disabled,
          'text-white bg-violet-400 border border-transparent shadow-sm enabled:bg-violet-600 enabled:hover:bg-violet-500 enabled:focus:ring-violet-600':
            variant === 'primary',
          'bg-white text-gray-500 border-violet-300 border shadow-sm enabled:hover:bg-gray-50 enabled:focus:ring-gray-500':
            variant === 'secondary',
          'text-violet-700 border-violet-300 enabled:hover:bg-violet-100 enabled:focus:ring-violet-500':
            variant === 'soft'
        }
      )}
      disabled={disabled}
      title={title}
    >
      {children}
    </button>
  );
}
