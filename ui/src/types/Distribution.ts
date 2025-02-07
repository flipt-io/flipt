export interface IDistribution {
  variant: string;
  rollout: number;
}

export enum DistributionType {
  None = 'none',
  Single = 'single',
  Multi = 'multi'
}
