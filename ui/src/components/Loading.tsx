import { cls } from '~/utils/helpers';

type LoadingProps = {
  isPrimary?: boolean;
  fullScreen?: boolean;
};

export default function Loading(props: LoadingProps) {
  const { isPrimary, fullScreen } = props;

  return (
    <div
      className={cls('flex items-center justify-center', {
        'h-screen': fullScreen
      })}
    >
      <div
        className={cls(
          'border-violet-300 h-5 w-5 animate-spin rounded-full border-b-2',
          {
            'border-white-300': isPrimary
          }
        )}
      ></div>
    </div>
  );
}
