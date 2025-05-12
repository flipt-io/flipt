import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

import { SidebarTrigger } from '~/components/ui/sidebar';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { Badge } from './Badge';

export function Header({
  ns,
  env,
  sidebarOpen
}: {
  ns: string;
  env: string;
  sidebarOpen: boolean;
}) {
  const info = useSelector(selectInfo);
  const topbarStyle = {
    backgroundColor: info?.ui?.topbarColor,
    borderRadius: '1rem 1rem 0 0'
  };
  return (
    <header
      className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
      style={topbarStyle}
    >
      <div className="flex w-full px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        {!sidebarOpen && (
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge variant="outlinemuted" className="ml-auto">
                {ns}@{env}
              </Badge>
            </TooltipTrigger>
            <TooltipContent side="bottom" align="end">
              Current namespace and environment
            </TooltipContent>
          </Tooltip>
        )}
      </div>
    </header>
  );
}
