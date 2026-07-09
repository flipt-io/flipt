import { IFlag } from '~/types/Flag';

export function flagReferencesSegment(
  flag: IFlag,
  segmentKey: string
): boolean {
  return (
    flag.rules?.some((rule) => rule.segments?.includes(segmentKey)) ||
    flag.rollouts?.some((rollout) =>
      rollout.segment?.segments?.includes(segmentKey)
    ) ||
    false
  );
}

export function flagsReferencingSegment(
  flags: IFlag[],
  segmentKey: string
): IFlag[] {
  return flags
    .filter((flag) => flagReferencesSegment(flag, segmentKey))
    .sort((a, b) => a.key.localeCompare(b.key));
}
