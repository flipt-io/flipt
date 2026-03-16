import { Search } from 'lucide-react';
import { useEffect, useState } from 'react';

type SearchboxProps = {
  value: string;
  onChange: (value: string) => void;
  debounce?: number;
  className?: string;
};

export default function Searchbox(props: SearchboxProps) {
  const {
    value: initialValue,
    onChange,
    debounce = 500,
    className = ''
  } = props;
  const [value, setValue] = useState<string>(initialValue);

  useEffect(() => {
    setValue(initialValue);
  }, [initialValue]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      onChange(value);
    }, debounce);

    return () => clearTimeout(timeout);
  }, [debounce, onChange, value]);

  return (
    <div className={`${className} flex flex-1 items-center justify-start`}>
      <div className="w-full max-w-60 lg:max-w-md">
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <Search
              className="text-muted-foreground h-4 w-4"
              aria-hidden="true"
            />
          </div>
          <input
            id="search"
            name="search"
            className="bg-input/30 text-input/90 border-input text-secondary-foreground placeholder-muted-foreground focus:ring-brand block h-8 w-full rounded-md border py-2 pr-3 pl-10 leading-4 shadow-xs focus:outline-hidden sm:text-sm"
            placeholder="Search"
            type="search"
            value={value}
            onChange={(e) => setValue(e.target.value)}
          />
        </div>
      </div>
    </div>
  );
}
