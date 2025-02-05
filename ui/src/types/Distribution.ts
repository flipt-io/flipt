export interface IDistributionBase {
  variant: string;
  rollout: number;
}

export interface IDistributionVariant extends IDistributionBase {
  variantId: string;
}

export interface IDistribution extends IDistributionBase {
  id: string;
}

export enum DistributionType {
  None = 'none',
  Single = 'single',
  Multi = 'multi'
}
