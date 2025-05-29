import { cn } from '~/components/utils';

export interface PageHeaderProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'> {
  title: React.ReactNode;
}

const PageHeader = ({ title, children, className }: PageHeaderProps) => (
  <div className={cn('flex items-center justify-between', className)}>
    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 sm:text-3xl sm:tracking-tight">
      {title}
    </h1>
    {children}
  </div>
);
PageHeader.displayName = 'PageHeader';

export { PageHeader };
