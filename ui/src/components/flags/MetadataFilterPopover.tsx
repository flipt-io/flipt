import { SlidersHorizontalIcon } from 'lucide-react';
import { useState } from 'react';

import { BaseInput } from '~/components/BaseInput';
import { Button } from '~/components/Button';
import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';

import { MetadataFilter } from '~/types/Flag';

interface MetadataFilterPopoverProps {
  availableKeys: string[];
  onAdd: (filter: MetadataFilter) => void;
}

export default function MetadataFilterPopover({
  availableKeys,
  onAdd
}: MetadataFilterPopoverProps) {
  const [open, setOpen] = useState(false);
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');

  const canAdd = key.trim().length > 0 && value.trim().length > 0;

  const handleAdd = () => {
    if (!canAdd) return;
    onAdd({ key: key.trim(), value: value.trim() });
    setKey('');
    setValue('');
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="secondaryline"
          aria-label="Filter"
          className="gap-2"
        >
          <SlidersHorizontalIcon className="h-4 w-4" />
          Filter
        </Button>
      </PopoverTrigger>

      <PopoverContent align="start" className="w-72 space-y-3">
        <p className="text-sm font-medium">Filter by metadata</p>

        {/* Key input with datalist for autocomplete */}
        <div>
          <label
            htmlFor="mf-key"
            className="text-xs text-muted-foreground mb-1 block"
          >
            Key
          </label>
          <BaseInput
            id="mf-key"
            type="text"
            placeholder="Key"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            list="mf-available-keys"
            className="w-full"
          />
          <datalist id="mf-available-keys">
            {availableKeys.map((k) => (
              <option key={k} value={k} />
            ))}
          </datalist>
        </div>

        {/* Value input */}
        <div>
          <label
            htmlFor="mf-value"
            className="text-xs text-muted-foreground mb-1 block"
          >
            Value
          </label>
          <BaseInput
            id="mf-value"
            type="text"
            placeholder="Value"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            className="w-full"
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleAdd();
            }}
          />
        </div>

        <Button
          type="button"
          variant="primary"
          className="w-full"
          disabled={!canAdd}
          onClick={handleAdd}
          aria-label="Add filter"
        >
          Add filter
        </Button>
      </PopoverContent>
    </Popover>
  );
}
