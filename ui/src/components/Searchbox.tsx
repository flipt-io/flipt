import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
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
    <div
      className={`${className} flex flex-1 items-center justify-center lg:justify-start`}
    >
      <div className="lg:max-w-s w-full max-w-lg">
        <label htmlFor="search" className="sr-only">
          Search
        </label>
        <div className="relative">
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
            <MagnifyingGlassIcon
              className="text-gray-400 h-5 w-5"
              aria-hidden="true"
            />
          </div>
          <input
            id="search"
            name="search"
            className="bg-white text-gray-900 border-gray-300 block w-full rounded-md border py-2 pl-10 pr-3 leading-5 placeholder-gray-500 shadow-sm focus:border-violet-400 focus:placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-violet-400 sm:text-sm"
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
