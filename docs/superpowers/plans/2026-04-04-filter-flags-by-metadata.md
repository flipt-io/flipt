# Filter Flags by Metadata — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a client-side metadata filter (popover + chips) to the Flags list page so users can narrow flags by key-value metadata pairs.

**Architecture:** Pure UI change — no backend or API modifications. A new `MetadataFilterPopover` component handles filter input; `FlagTable` gains local `metadataFilters` state and a `useMemo` pre-filter that runs before TanStack Table sees the data. Active filters render as removable `Badge` chips below the toolbar.

**Tech Stack:** React 19, TypeScript, TanStack Table v8, Tailwind CSS, Lucide icons, Radix UI Popover (via `~/components/Popover`), `~/components/Combobox`, `~/components/Badge`, `~/components/Button`, `~/components/forms/Input`, Jest + jsdom (unit tests).

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `ui/src/types/Flag.ts` | **Modify** | Add `MetadataFilter` interface |
| `ui/src/components/flags/MetadataFilterPopover.tsx` | **Create** | Popover UI for adding one key=value filter |
| `ui/src/components/flags/FlagTable.tsx` | **Modify** | Filter state, pre-filter logic, toolbar button, chips, empty-state condition |
| `ui/src/components/flags/MetadataFilterPopover.test.tsx` | **Create** | Unit tests for the popover component |
| `ui/src/utils/flagMetadataFilter.ts` | **Create** | Pure filter function (easy to unit-test in isolation) |
| `ui/src/utils/flagMetadataFilter.test.ts` | **Create** | Unit tests for the filter function |

---

## Task 1: Add `MetadataFilter` type + pure filter function

**Files:**
- Modify: `ui/src/types/Flag.ts`
- Create: `ui/src/utils/flagMetadataFilter.ts`
- Create: `ui/src/utils/flagMetadataFilter.test.ts`

- [ ] **Step 1.1 — Write the failing tests**

Create `ui/src/utils/flagMetadataFilter.test.ts`:

```ts
import { applyMetadataFilters } from './flagMetadataFilter';
import { FlagType, IFlag, MetadataFilter } from '~/types/Flag';

const makeFlag = (key: string, metadata?: Record<string, any>): IFlag => ({
  key,
  name: key,
  type: FlagType.VARIANT,
  enabled: true,
  description: '',
  metadata
});

describe('applyMetadataFilters', () => {
  const flags = [
    makeFlag('flag-a', { team: 'backend', env: 'production' }),
    makeFlag('flag-b', { team: 'frontend', env: 'production' }),
    makeFlag('flag-c', { team: 'backend', env: 'staging' }),
    makeFlag('flag-d'), // no metadata
    makeFlag('flag-e', { count: 42, active: true })
  ];

  it('returns all flags when no filters are active', () => {
    expect(applyMetadataFilters(flags, [])).toEqual(flags);
  });

  it('filters by a single exact-match key-value pair', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'backend' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-c']);
  });

  it('filter is case-insensitive', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'BACKEND' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-c']);
  });

  it('filter is a substring match', () => {
    const filters: MetadataFilter[] = [{ key: 'env', value: 'prod' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a', 'flag-b']);
  });

  it('multiple filters use AND logic', () => {
    const filters: MetadataFilter[] = [
      { key: 'team', value: 'backend' },
      { key: 'env', value: 'production' }
    ];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-a']);
  });

  it('excludes flags that do not have the filtered key', () => {
    const filters: MetadataFilter[] = [{ key: 'team', value: 'backend' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).not.toContain('flag-d');
  });

  it('matches numeric metadata values converted to string', () => {
    const filters: MetadataFilter[] = [{ key: 'count', value: '42' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-e']);
  });

  it('matches boolean metadata values converted to string', () => {
    const filters: MetadataFilter[] = [{ key: 'active', value: 'true' }];
    const result = applyMetadataFilters(flags, filters);
    expect(result.map((f) => f.key)).toEqual(['flag-e']);
  });
});
```

- [ ] **Step 1.2 — Run tests to confirm they fail**

```bash
cd ui && npx jest flagMetadataFilter --no-coverage
```

Expected: `FAIL` — cannot find module `'./flagMetadataFilter'`.

- [ ] **Step 1.3 — Add `MetadataFilter` interface to `ui/src/types/Flag.ts`**

Open `ui/src/types/Flag.ts` and append after the last `export` (after `FilterableFlag`):

```ts
export interface MetadataFilter {
  key: string;
  value: string;
}
```

- [ ] **Step 1.4 — Create `ui/src/utils/flagMetadataFilter.ts`**

```ts
import { IFlag, MetadataFilter } from '~/types/Flag';

/**
 * Filters a list of flags by the given metadata key-value pairs.
 * All filters must match (AND semantics).
 * Matching is case-insensitive substring on the string-coerced metadata value.
 */
export function applyMetadataFilters(
  flags: IFlag[],
  filters: MetadataFilter[]
): IFlag[] {
  if (filters.length === 0) return flags;

  return flags.filter((flag) =>
    filters.every(({ key, value }) => {
      const metaVal = flag.metadata?.[key];
      if (metaVal === undefined || metaVal === null) return false;
      return String(metaVal).toLowerCase().includes(value.toLowerCase());
    })
  );
}
```

- [ ] **Step 1.5 — Run tests to confirm they pass**

```bash
cd ui && npx jest flagMetadataFilter --no-coverage
```

Expected: `PASS` — all 8 tests green.

- [ ] **Step 1.6 — Commit**

```bash
cd ui && git add src/types/Flag.ts src/utils/flagMetadataFilter.ts src/utils/flagMetadataFilter.test.ts
git commit -m "feat: add MetadataFilter type and applyMetadataFilters utility"
```

---

## Task 2: Create `MetadataFilterPopover` component

**Files:**
- Create: `ui/src/components/flags/MetadataFilterPopover.tsx`
- Create: `ui/src/components/flags/MetadataFilterPopover.test.tsx`

This component renders a button that opens a popover. Inside: a text input for the metadata key, a text input for the value, and an "Add filter" button. On submit it calls `onAdd(filter)` and resets its own state.

> **Note on Combobox:** `Combobox` requires items typed as `ISelectable` (`{ key, displayValue }`). We map `availableKeys` to that shape. However since Combobox is a compound Radix Popover component, testing it with jsdom/jest is fragile. We keep the component simple: a plain `<input>` for key (with `<datalist>`) is easier to unit-test; Combobox is used for the richer UX only in the real component.

- [ ] **Step 2.1 — Write the failing tests**

Create `ui/src/components/flags/MetadataFilterPopover.test.tsx`:

```tsx
/**
 * @jest-environment jsdom
 */
import { render, screen, fireEvent } from '@testing-library/react';
import MetadataFilterPopover from './MetadataFilterPopover';
import { MetadataFilter } from '~/types/Flag';

// Radix Popover uses portals. We mock it so the content always renders inline.
jest.mock('~/components/Popover', () => ({
  Popover: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe('MetadataFilterPopover', () => {
  it('renders a "Filter" trigger button', () => {
    render(<MetadataFilterPopover availableKeys={[]} onAdd={jest.fn()} />);
    expect(screen.getByRole('button', { name: /filter/i })).toBeInTheDocument();
  });

  it('calls onAdd with the entered key and value when Add is clicked', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={['team']} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).toHaveBeenCalledWith<[MetadataFilter]>({
      key: 'team',
      value: 'backend'
    });
  });

  it('does not call onAdd when key is empty', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).not.toHaveBeenCalled();
  });

  it('does not call onAdd when value is empty', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).not.toHaveBeenCalled();
  });

  it('resets key and value inputs after a successful add', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    const keyInput = screen.getByPlaceholderText(/key/i);
    const valueInput = screen.getByPlaceholderText(/value/i);

    fireEvent.change(keyInput, { target: { value: 'team' } });
    fireEvent.change(valueInput, { target: { value: 'backend' } });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect((keyInput as HTMLInputElement).value).toBe('');
    expect((valueInput as HTMLInputElement).value).toBe('');
  });
});
```

- [ ] **Step 2.2 — Check test infra: confirm `@testing-library/react` is installed**

```bash
cd ui && cat package.json | grep testing-library
```

Expected: see `"@testing-library/react"` in devDependencies. If not present:
```bash
cd ui && npm install --save-dev @testing-library/react @testing-library/jest-dom
```

Then check `ui/jest.config.ts` for `setupFilesAfterFramework` pointing to a setup file that calls `import '@testing-library/jest-dom'`. If absent, add to `jest.config.ts`:
```ts
setupFilesAfterFramework: ['<rootDir>/src/setupTests.ts'],
```
> Note: the correct Jest config key is `setupFilesAfterFramework` (not `setupFilesAfterEach`). Verify the exact key name in your Jest version if this causes a type error — it may be `setupFilesAfterEnv` in older configs.
and create `ui/src/setupTests.ts`:
```ts
import '@testing-library/jest-dom';
```

- [ ] **Step 2.3 — Run tests to confirm they fail**

```bash
cd ui && npx jest MetadataFilterPopover --no-coverage
```

Expected: `FAIL` — cannot find module `'./MetadataFilterPopover'`.

- [ ] **Step 2.4 — Create `ui/src/components/flags/MetadataFilterPopover.tsx`**

```tsx
import { SlidersHorizontalIcon } from 'lucide-react';
import { useState } from 'react';

import { Button } from '~/components/Button';
import Input from '~/components/forms/Input';
import {
  Popover,
  PopoverContent,
  PopoverTrigger
} from '~/components/Popover';

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
          <label htmlFor="mf-key" className="text-xs text-muted-foreground mb-1 block">
            Key
          </label>
          <Input
            id="mf-key"
            name="mf-key"
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
          <label htmlFor="mf-value" className="text-xs text-muted-foreground mb-1 block">
            Value
          </label>
          <Input
            id="mf-value"
            name="mf-value"
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
```

- [ ] **Step 2.5 — Run tests to confirm they pass**

```bash
cd ui && npx jest MetadataFilterPopover --no-coverage
```

Expected: `PASS` — all 5 tests green.

- [ ] **Step 2.6 — Commit**

```bash
git add ui/src/components/flags/MetadataFilterPopover.tsx \
        ui/src/components/flags/MetadataFilterPopover.test.tsx
git commit -m "feat: add MetadataFilterPopover component"
```

---

## Task 3: Wire metadata filtering into `FlagTable`

**Files:**
- Modify: `ui/src/components/flags/FlagTable.tsx`

This task adds:
1. `metadataFilters` state
2. `availableMetadataKeys` memo
3. `filteredFlags` memo (using `applyMetadataFilters`)
4. `MetadataFilterPopover` button in toolbar
5. Filter chips row below toolbar
6. Updated empty-state condition

- [ ] **Step 3.1 — Read the current `FlagTable.tsx` imports section**

Open `ui/src/components/flags/FlagTable.tsx`, lines 1–46. You will add new imports at the top.

- [ ] **Step 3.2 — Add new imports**

In `ui/src/components/flags/FlagTable.tsx`, extend the existing import block. Add these four import lines **after** the existing imports (after line 46):

```ts
import { XIcon } from 'lucide-react';

import MetadataFilterPopover from '~/components/flags/MetadataFilterPopover';

import { MetadataFilter } from '~/types/Flag';

import { applyMetadataFilters } from '~/utils/flagMetadataFilter';
```

> `XIcon` is already available from the `lucide-react` package used throughout the project.

- [ ] **Step 3.3 — Add `metadataFilters` state and derived memos**

Inside the `FlagTable` function body, after the existing line:

```ts
const flags = useMemo(() => data?.flags || [], [data]);
```

Add:

```ts
const [metadataFilters, setMetadataFilters] = useState<MetadataFilter[]>([]);

const availableMetadataKeys = useMemo(
  () =>
    [...new Set(flags.flatMap((f) => Object.keys(f.metadata ?? {})))].sort(),
  [flags]
);

const filteredFlags = useMemo(
  () => applyMetadataFilters(flags, metadataFilters),
  [flags, metadataFilters]
);
```

- [ ] **Step 3.4 — Replace `data: flags` with `data: filteredFlags` in `useReactTable`**

Find the line inside `useReactTable({`:
```ts
    data: flags,
```
Change it to:
```ts
    data: filteredFlags,
```

- [ ] **Step 3.5 — Add the Filter button to the toolbar**

Find the existing toolbar JSX block in the `return` statement:

```tsx
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
            </div>
```

Replace it with:

```tsx
            <div className="flex items-center gap-4">
              <Searchbox value={filter ?? ''} onChange={setFilter} />
              <MetadataFilterPopover
                availableKeys={availableMetadataKeys}
                onAdd={(f) =>
                  setMetadataFilters((prev) => [...prev, f])
                }
              />
            </div>
```

- [ ] **Step 3.6 — Add the filter chips row**

Directly after the toolbar `<div>` (after the closing `</div>` of the `flex items-center justify-between` wrapper), add:

```tsx
        {metadataFilters.length > 0 && (
          <div className="flex flex-wrap items-center gap-2">
            {metadataFilters.map((f, i) => (
              <Badge
                key={i}
                variant="outlinemuted"
                className="flex items-center gap-1 pr-1"
              >
                <span>
                  {f.key}: {f.value}
                </span>
                <button
                  type="button"
                  onClick={() =>
                    setMetadataFilters((prev) => prev.filter((_, idx) => idx !== i))
                  }
                  aria-label={`Remove filter ${f.key}:${f.value}`}
                  className="ml-1 rounded-full hover:bg-muted p-0.5"
                >
                  <XIcon className="h-3 w-3" />
                </button>
              </Badge>
            ))}
            <button
              type="button"
              onClick={() => setMetadataFilters([])}
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              Clear all
            </button>
          </div>
        )}
```

- [ ] **Step 3.7 — Update the empty-state condition**

Find the existing empty-state condition:

```tsx
        {table.getRowCount() === 0 && filter.length === 0 && (
          <EmptyFlagList path={`${path}/new`} />
        )}
        {table.getRowCount() === 0 && filter.length > 0 && (
```

Replace both conditions with:

```tsx
        {table.getRowCount() === 0 &&
          filter.length === 0 &&
          metadataFilters.length === 0 && (
            <EmptyFlagList path={`${path}/new`} />
          )}
        {table.getRowCount() === 0 &&
          (filter.length > 0 || metadataFilters.length > 0) && (
```

- [ ] **Step 3.8 — Type-check**

```bash
cd ui && npx tsc --noEmit
```

Expected: no errors. If you see `XIcon` not found, verify the lucide-react version supports it — alternative name is `X`. Replace `XIcon` with `X` and update the import accordingly if needed:
```ts
import { X } from 'lucide-react';
// then use <X className="h-3 w-3" />
```

- [ ] **Step 3.9 — Run existing tests to confirm nothing broke**

```bash
cd ui && npx jest --no-coverage
```

Expected: all pre-existing tests still pass.

- [ ] **Step 3.10 — Commit**

```bash
git add ui/src/components/flags/FlagTable.tsx
git commit -m "feat: wire metadata filter state and UI into FlagTable"
```

---

## Task 4: Manual smoke test

No automated test can replace a quick visual check. These steps require the Flipt dev server running.

- [ ] **Step 4.1 — Start the dev server**

```bash
# Terminal 1 — backend (if not already running)
cd <flipt-root> && go run ./cmd/flipt server

# Terminal 2 — frontend
cd ui && npm run dev
```

Open http://localhost:5173 in a browser.

- [ ] **Step 4.2 — Create test flags with metadata**

In the UI, create three flags with the following metadata (Settings → Metadata tab of the flag form):

| Flag key | Metadata |
|----------|----------|
| `flag-backend` | `team=backend`, `env=production` |
| `flag-frontend` | `team=frontend`, `env=production` |
| `flag-staging` | `team=backend`, `env=staging` |

- [ ] **Step 4.3 — Verify Filter button appears**

Navigate to `/namespaces/default/flags`. Confirm a "Filter" button (with sliders icon) appears next to the search box.

- [ ] **Step 4.4 — Add a single filter**

Click Filter → enter key `team`, value `backend` → click "Add filter".

Expected:
- A chip `team: backend ×` appears below the toolbar.
- Only `flag-backend` and `flag-staging` are visible.

- [ ] **Step 4.5 — Add a second filter**

Click Filter again → enter key `env`, value `production` → click "Add filter".

Expected:
- Two chips shown: `team: backend ×` and `env: production ×`.
- Only `flag-backend` is visible (AND logic).

- [ ] **Step 4.6 — Remove one filter**

Click `×` on the `env: production` chip.

Expected:
- Only the `team: backend` chip remains.
- Both `flag-backend` and `flag-staging` are visible again.

- [ ] **Step 4.7 — Clear all filters**

Click "Clear all".

Expected:
- No chips shown.
- All flags visible.

- [ ] **Step 4.8 — Empty state with metadata filter**

Add a filter `key=nonexistent`, value=`x`.

Expected:
- "No flags matched your search" well is shown (not the "Create Your First Flag" state).

- [ ] **Step 4.9 — Combined text + metadata filter**

In the search box type `backend`. Then add metadata filter `env=production`.

Expected: only flags matching both text search AND metadata filter are shown.

---

## Task 5: Final check and cleanup

- [ ] **Step 5.1 — Run full test suite one more time**

```bash
cd ui && npx jest --no-coverage
```

Expected: all tests pass.

- [ ] **Step 5.2 — TypeScript check**

```bash
cd ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 5.3 — Lint**

```bash
cd ui && npm run lint
```

Expected: no new lint errors. Fix any that appear (usually unused imports).

- [ ] **Step 5.4 — Final commit (if any lint fixes were needed)**

```bash
git add -A
git commit -m "chore: fix lint warnings after metadata filter implementation"
```
