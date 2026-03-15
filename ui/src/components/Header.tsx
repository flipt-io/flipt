import { useSelector } from 'react-redux';
import { selectConfig, selectReadonly } from '~/app/meta/metaSlice';
import { SidebarTrigger } from '~/components/Sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '~/components/Tooltip';
import { Badge } from '~/components/Badge';
import ReadOnly from './header/ReadOnly';

export function Header({ ns }: { ns: string }) {
  const readOnly = useSelector(selectReadonly);
  const { ui } = useSelector(selectConfig);
  const topbarStyle = { backgroundColor: ui.topbar?.color };

  return (
    <header
      className="bg-background flex h-12 shrink-0 items-center gap-2 border-b px-4 lg:gap-2 lg:px-6"
      style={topbarStyle}
    >
      <SidebarTrigger className="text-muted-foreground flex-shrink-0" />
      <div className="flex-1" />
      <div className="flex flex-shrink-0 items-center gap-2">
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge
              variant="secondary"
              className="flex-shrink-0 cursor-pointer px-3 py-1 text-xs"
            >
              {ns}
            </Badge>
          </TooltipTrigger>
          <TooltipContent side="bottom" align="end">
            Current namespace
          </TooltipContent>
        </Tooltip>

        {readOnly && <ReadOnly />}
      </div>
    </header>
  );
}
