import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectInfo } from '~/app/meta/metaSlice';

import { SidebarTrigger } from '~/components/ui/sidebar';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { Badge } from './Badge';
import { EnvironmentRemoteInfo } from './environments/EnvironmentRemoteInfo';

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
  const currentEnvironment = useSelector(selectCurrentEnvironment);
  const topbarStyle = {
    backgroundColor: info?.ui?.topbarColor,
    borderRadius: '1rem 1rem 0 0'
  };
  return (
    <header
      className="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
      style={topbarStyle}
    >
      <div className="flex w-full items-center justify-between px-4 lg:gap-2 lg:px-6">
        <div className="flex items-center justify-between w-full gap-3 min-w-0">
          <SidebarTrigger className="-ml-2" />
          <EnvironmentRemoteInfo environment={currentEnvironment} />
        </div>
        {!sidebarOpen && (
          <>
            {currentEnvironment?.configuration?.remote && (
              <span className="mx-2 h-6 w-px bg-border rounded" />
            )}
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge
                  variant="secondary"
                  className="ml-auto px-3 py-1 bg-background font-semibold text-xs"
                >
                  {ns}@{env}
                </Badge>
              </TooltipTrigger>
              <TooltipContent side="bottom" align="end">
                Current namespace and environment
              </TooltipContent>
            </Tooltip>
          </>
        )}
      </div>
    </header>
  );
}
