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
