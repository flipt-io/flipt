import { cn } from '~/lib/utils';

export interface PageHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
}

const PageHeader = ({ title, children, className }: PageHeaderProps) => (
  <div className={cn('flex items-center justify-between', className)}>
    <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
      {title}
    </h1>
    {children}
  </div>
);
PageHeader.displayName = 'PageHeader';

export { PageHeader };
