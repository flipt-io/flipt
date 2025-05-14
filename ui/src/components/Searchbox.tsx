import { SearchIcon } from 'lucide-react';
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
      <div className="w-full">
        <label htmlFor="search" className="sr-only">
          Search
        </label>
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <SearchIcon
              className="h-5 w-5 text-gray-400 dark:text-gray-500"
              aria-hidden="true"
            />
          </div>
          <input
            id="search"
            name="search"
            className="block rounded-md border border-gray-300 dark:border-gray-700 bg-background dark:bg-gray-900 py-2 pl-10 pr-3 leading-5 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 shadow-xs sm:text-sm"
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
