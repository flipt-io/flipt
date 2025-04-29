import { useSelector } from 'react-redux';
import { Toaster as Sonner, ToasterProps } from 'sonner';

import { selectTheme } from '~/app/preferences/preferencesSlice';

const Toaster = ({ ...props }: ToasterProps) => {
  const theme = useSelector(selectTheme);
  return (
    <Sonner
      theme={theme as ToasterProps['theme']}
      className="toaster group"
      style={
        {
          '--normal-bg': 'var(--popover)',
          '--normal-text': 'var(--popover-foreground)',
          '--normal-border': 'var(--border)'
        } as React.CSSProperties
      }
      {...props}
    />
  );
};

export { Toaster };
