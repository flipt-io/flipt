import { cls } from '~/utils/helpers';

type PrecentProps = {
  className?: string;
};

export default function Percent(props: PrecentProps) {
  const { className } = props;

  return (
    <div
      className={cls(
        'pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-sm text-muted-foreground',
        className
      )}
    >
      %
    </div>
  );
}
