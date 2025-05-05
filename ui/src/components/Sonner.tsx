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
          '--normal-border': 'var(--border)',
          '--toast-bg': 'var(--popover)',
          '--toast-text': 'var(--popover-foreground)',
          '--toast-border': 'var(--border)',
          '--toast-box-shadow': '0 2px 5px rgba(0, 0, 0, 0.2)'
        } as React.CSSProperties
      }
      {...props}
    />
  );
};

export { Toaster };
