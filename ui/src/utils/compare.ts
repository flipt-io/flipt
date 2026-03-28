export type CompareStatus =
  | 'same'
  | 'different'
  | 'source_only'
  | 'target_only';

export function stableValue(value: unknown): string {
  if (value === null || typeof value !== 'object') {
    return JSON.stringify(value);
  }

  if (Array.isArray(value)) {
    return `[${value.map((v) => stableValue(v)).join(',')}]`;
  }

  return `{${Object.entries(value as Record<string, unknown>)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([k, v]) => `${k}:${stableValue(v)}`)
    .join(',')}}`;
}

export function areDifferent(a: unknown, b: unknown): boolean {
  return stableValue(a) !== stableValue(b);
}

export function diffFlag(
  source: Record<string, unknown>,
  target: Record<string, unknown>
): string[] {
  const differences: string[] = [];

  if (areDifferent(source.enabled, target.enabled)) {
    differences.push('enabled');
  }
  if (areDifferent(source.defaultVariant, target.defaultVariant)) {
    differences.push('default variant');
  }
  if (areDifferent(source.variants, target.variants)) {
    differences.push('variants');
  }
  if (areDifferent(source.rules, target.rules)) {
    differences.push('rules');
  }
  if (areDifferent(source.rollouts, target.rollouts)) {
    differences.push('rollouts');
  }

  return differences;
}

// Drift = structural differences only (excludes enabled/disabled state).
export function driftFields(
  source: Record<string, unknown>,
  target: Record<string, unknown>
): string[] {
  const fields: string[] = [];

  if (areDifferent(source.defaultVariant, target.defaultVariant)) {
    fields.push('default variant');
  }
  if (areDifferent(source.variants, target.variants)) {
    fields.push('variants');
  }
  if (areDifferent(source.rules, target.rules)) {
    fields.push('rules');
  }
  if (areDifferent(source.rollouts, target.rollouts)) {
    fields.push('rollouts');
  }

  return fields;
}

export function diffSegment(
  source: Record<string, unknown>,
  target: Record<string, unknown>
): string[] {
  const differences: string[] = [];

  if (areDifferent(source.matchType, target.matchType)) {
    differences.push('match type');
  }
  if (areDifferent(source.constraints, target.constraints)) {
    differences.push('constraints');
  }

  return differences;
}

export function toStatus(value: string): CompareStatus {
  switch (value) {
    case 'COMPARE_STATUS_DIFFERENT':
      return 'different';
    case 'COMPARE_STATUS_SOURCE_ONLY':
      return 'source_only';
    case 'COMPARE_STATUS_TARGET_ONLY':
      return 'target_only';
    default:
      return 'same';
  }
}

export function toResourceType(typeUrl: string): 'flag' | 'segment' | null {
  if (typeUrl === 'flipt.core.Flag') {
    return 'flag';
  }
  if (typeUrl === 'flipt.core.Segment') {
    return 'segment';
  }
  return null;
}
