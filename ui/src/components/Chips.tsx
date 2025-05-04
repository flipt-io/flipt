import { useState } from 'react';

export default function ChipList({
  values,
  maxItemCount = 5,
  showAll = false
}: {
  values: (string | number)[];
  maxItemCount?: number;
  showAll?: boolean;
}) {
  const [showAllValues, setShowAllValues] = useState<boolean>(showAll);

  const visibleValues = showAllValues ? values : values.slice(0, maxItemCount);
  return (
    <div className="flex flex-wrap gap-2">
      {visibleValues.map((value, i) => (
        <div
          className="max-w-32 truncate rounded bg-gray-100 dark:bg-gray-800 px-2 py-1 text-gray-900 dark:text-white"
          key={i}
        >
          {value}
        </div>
      ))}
      {!showAll && values.length > maxItemCount && (
        <button
          className="rounded px-1.5 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
          onClick={() => setShowAllValues((b) => !b)}
        >
          {showAllValues ? 'show less' : '...'}
        </button>
      )}
    </div>
  );
}
