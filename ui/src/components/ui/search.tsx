import { Input } from '~/components/ui/input';
import { Search as Icon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { cn } from '~/lib/utils';

type Props = {
  value: string;
  onChange: (value: string) => void;
  debounce?: number;
  className?: string;
};

const Search = ({
  value: initialValue,
  onChange,
  debounce = 200,
  className
}: Props) => {
  const [value, setValue] = useState<string>(initialValue);

  useEffect(() => {
    const timeout = setTimeout(() => {
      onChange(value);
    }, debounce);

    return () => clearTimeout(timeout);
  }, [debounce, onChange, value]);

  return (
    <div className="bg-background/95 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="relative">
        <Icon className="absolute left-2 top-2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search..."
          className={cn('pl-8', className)}
          type="search"
          value={value ?? ''}
          onChange={(e) => setValue(e.target.value)}
        />
      </div>
    </div>
  );
};

Search.displayName = 'Search';

export { Search };
