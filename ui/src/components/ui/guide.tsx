import React from 'react';
import { cn } from '~/lib/utils';
import Well from '~/components/Well';

export interface GuideProps extends React.HTMLAttributes<HTMLDivElement> {}

const Guide = ({ children, className, ...props }: GuideProps) => (
  <div className={cn('flex flex-col text-center', className)}>
    <Well>
      <p className="p-2 text-sm text-muted-foreground" {...props}>
        {children}
      </p>
    </Well>
  </div>
);

Guide.displayName = 'Guide';

export default Guide;
