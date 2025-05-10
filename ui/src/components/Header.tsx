import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

import { Separator } from '~/components/ui/separator';
import { SidebarTrigger } from '~/components/ui/sidebar';

export function Header({ ns }: { ns: string }) {
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
        <Separator
          orientation="vertical"
          className="mr-2 data-[orientation=vertical]:h-6 items-center"
        />
        <h1 className="text-base font-medium">{ns}</h1>
      </div>
    </header>
  );
}
