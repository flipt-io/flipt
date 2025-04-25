import { cls } from '~/utils/helpers';

export type Tag = {
  label: string;
  variant?: 'default' | 'outline' | 'purple' | 'blue' | 'green' | 'pink';
};

export function TagBadge({ tag }: { tag: Tag }) {
  const variants = {
    default: 'bg-secondary/50 text-secondary-foreground',
    outline: 'border border-border bg-transparent',
    purple: 'bg-violet-500/10 text-violet-500 dark:bg-violet-500/20',
    blue: 'bg-blue-500/10 text-blue-500 dark:bg-blue-500/20',
    green: 'bg-green-500/10 text-green-500 dark:bg-green-500/20',
    pink: 'bg-pink-500/10 text-pink-500 dark:bg-pink-500/20'
  };

  return (
    <span
      className={cls(
        'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium',
        variants[tag.variant || 'default']
      )}
    >
      {tag.label}
    </span>
  );
}
