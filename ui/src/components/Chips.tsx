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
          className="text-secondary-foreground max-w-32 truncate rounded-sm px-1.5 py-0.5"
          key={i}
        >
          {value}
        </div>
      ))}
      {!showAll && values.length > maxItemCount && (
        <button
          className="text-secondary-foreground rounded-sm px-1.5"
          onClick={() => setShowAllValues((b) => !b)}
        >
          {showAllValues ? 'show less' : '...'}
        </button>
      )}
    </div>
  );
}
