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
          className="text-gray-900 bg-gray-200 max-w-32 truncate rounded px-1.5 py-0.5"
          key={i}
        >
          {value}
        </div>
      ))}
      {!showAll && values.length > maxItemCount && (
        <button
          className="text-gray-700 rounded px-1.5"
          onClick={() => setShowAllValues((b) => !b)}
        >
          {showAllValues ? 'show less' : '...'}
        </button>
      )}
    </div>
  );
}
