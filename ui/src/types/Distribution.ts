export interface IDistributionBase {
  variantId: string;
  rollout: number;
}

export interface IDistributionVariant extends IDistributionBase {
  variantKey: string;
}

export interface IDistribution extends IDistributionBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export enum DistributionType {
  None = 'none',
  Single = 'single',
  Multi = 'multi'
}
