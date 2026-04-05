# Filter Flags by Metadata — Design Spec

**Issue**: https://github.com/flipt-io/flipt/issues/3739  
**Date**: 2026-04-04  
**Scope**: UI only (client-side), no backend changes required  

---

## Problem

The Flags list page (`/namespaces/:ns/flags`) only supports text search over name, key, and description. Users who tag flags with metadata key-value pairs (e.g., `team: backend`, `env: production`) have no way to filter the list by those tags. This makes managing large flag sets difficult.

---

## Goals

- Allow users to filter flags by one or more metadata key-value pairs.
- Multiple filters combine with AND logic (flag must match all active filters).
- No backend or API changes required.
- Consistent with existing Flipt UI patterns (Combobox, Badge, Popover).

## Non-Goals

- Server-side filtering via API query params (future enhancement).
- Special-casing a `tags` metadata key — all metadata keys are treated equally.
- Persisting filters across page navigation or sessions.

---

## Architecture

All changes are confined to `ui/src/`:

```
ui/src/
  components/flags/
    FlagTable.tsx          ← add filter state, pre-filter logic, toolbar chips
    MetadataFilterPopover.tsx   ← new component: popover for adding a filter
  types/
    Flag.ts                ← add MetadataFilter interface (no existing types changed)
```

No changes to:
- Backend Go code
- Protobuf definitions
- `flagsApi.ts` API layer
- Any other component outside `components/flags/`

---

## Data Model

```ts
// ui/src/types/Flag.ts — new export
export interface MetadataFilter {
  key: string;   // metadata key, e.g. "team"
  value: string; // partial match value, e.g. "backend"
}
```

---

## Filter Logic

Pre-filter the flags array before passing it to `useReactTable`. This avoids hooking into TanStack Table's filter API and keeps the logic simple.

```ts
const filteredFlags = useMemo(() => {
  if (metadataFilters.length === 0) return flags;
  return flags.filter((flag) =>
    metadataFilters.every(({ key, value }) => {
      const metaVal = flag.metadata?.[key];
      if (metaVal === undefined) return false;
      return String(metaVal).toLowerCase().includes(value.toLowerCase());
    })
  );
}, [flags, metadataFilters]);
```

- Match is **case-insensitive substring** (consistent with the existing text search behaviour).
- A flag without the metadata key does **not** match a filter on that key.
- All filters must match (AND semantics). OR is out of scope.

---

## Component Design

### `MetadataFilterPopover`

A self-contained popover that lets the user pick a metadata key and enter a value, then emits an `onAdd(filter: MetadataFilter)` callback.

Props:
```ts
interface MetadataFilterPopoverProps {
  availableKeys: string[];          // deduplicated keys from all current flags
  onAdd: (filter: MetadataFilter) => void;
}
```

Internals:
- **Key input**: `Combobox` (existing component) populated with `availableKeys`. Also accepts free-text for keys not yet present in any flag.
- **Value input**: plain `<Input>` text field.
- **Add button**: disabled until both key and value are non-empty. Closes the popover on success.
- Uses the existing `Popover` / `PopoverContent` primitives from `~/components/`.

### Changes to `FlagTable`

**State additions:**
```ts
const [metadataFilters, setMetadataFilters] = useState<MetadataFilter[]>([]);
```

**Available keys** (memoized):
```ts
const availableMetadataKeys = useMemo(() =>
  [...new Set(flags.flatMap((f) => Object.keys(f.metadata ?? {})))].sort(),
  [flags]
);
```

**Toolbar** — between `<Searchbox>` and `<DataTableViewOptions>`:
```tsx
<MetadataFilterPopover
  availableKeys={availableMetadataKeys}
  onAdd={(f) => setMetadataFilters((prev) => [...prev, f])}
/>
```

**Filter chips** — rendered below the toolbar when `metadataFilters.length > 0`:
```tsx
{metadataFilters.length > 0 && (
  <div className="flex flex-wrap gap-2 items-center">
    {metadataFilters.map((f, i) => (
      <Badge key={i} variant="outlinemuted" className="gap-1">
        {f.key}: {f.value}
        <button onClick={() => removeFilter(i)} aria-label={`Remove filter ${f.key}`}>
          <XIcon className="h-3 w-3" />
        </button>
      </Badge>
    ))}
    <button
      className="text-xs text-muted-foreground hover:text-foreground"
      onClick={() => setMetadataFilters([])}
    >
      Clear all
    </button>
  </div>
)}
```

**Pass pre-filtered data to `useReactTable`**:
```ts
// Replace `data: flags` with `data: filteredFlags`
const table = useReactTable({
  data: filteredFlags,
  ...
});
```

---

## UI Layout

Before (current toolbar):
```
[🔍 Search...]                          [View ▼]
```

After (no active filters):
```
[🔍 Search...]  [⊕ Filter ▼]           [View ▼]
```

After (with active filters):
```
[🔍 Search...]  [⊕ Filter ▼]           [View ▼]
[×team: backend]  [×env: prod]  Clear all
```

The Filter button uses a `SlidersHorizontal` or `FilterIcon` icon from Lucide (consistent with existing `DataTableViewOptions`).

---

## Empty State

When no flags match the active filters:
- The existing "No flags matched your search" `<Well>` is reused — no new empty state needed.
- The condition that triggers this state must be updated: currently `filter.length > 0`, it must become `filter.length > 0 || metadataFilters.length > 0` so metadata-only filtering also shows this message instead of the "empty namespace" state.

---

## Edge Cases

| Case | Behaviour |
|------|-----------|
| Flag has no `metadata` field | Does not match any metadata filter |
| Metadata value is a number or boolean | Converted to string for comparison (`String(42)` → `"42"`) |
| Same key added twice with different values | Both filters apply (AND); effectively narrows to flags matching both values for that key — unlikely to find results, but not prevented |
| Metadata values that are objects or arrays | `String({...})` produces `"[object Object]"` — acceptable edge case; not filtered specially |
| Empty `availableKeys` (all flags have no metadata) | Combobox shows empty list; user can still type a key free-form |
| Text search + metadata filters both active | Both apply independently; flags must satisfy both to appear |

---

## Testing

- Unit tests for the filter logic function (pure function, easy to test).
- Component tests for `MetadataFilterPopover`: renders, adds a filter, clears input.
- Component tests for `FlagTable` active filter chips: add filter → chip appears; remove chip → filter gone; clear all → all chips gone.
- Existing `FlagTable` tests must continue to pass.

---

## Files Changed

| File | Change |
|------|--------|
| `ui/src/types/Flag.ts` | Add `MetadataFilter` interface |
| `ui/src/components/flags/FlagTable.tsx` | Add state, pre-filter logic, toolbar button, chips |
| `ui/src/components/flags/MetadataFilterPopover.tsx` | New component |
| `ui/src/components/flags/FlagTable.test.tsx` | New/updated tests |
| `ui/src/components/flags/MetadataFilterPopover.test.tsx` | New tests |
