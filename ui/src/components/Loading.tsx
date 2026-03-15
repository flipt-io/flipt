import { Loader2 } from 'lucide-react';
import { cls } from '~/utils/helpers';

type LoadingProps = {
  isPrimary?: boolean;
  fullScreen?: boolean;
};

export default function Loading(props: LoadingProps) {
  const { fullScreen } = props;

  return (
    <div
      className={cls('flex items-center justify-center', {
        'h-screen': fullScreen
      })}
    >
      <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
    </div>
  );
}
