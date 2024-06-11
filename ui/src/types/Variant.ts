import { IFilterable } from './Selectable';

export interface IVariantBase {
  key: string;
  name: string;
  description: string;
  attachment?: string;
  default?: boolean;
}

export interface IVariant extends IVariantBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export type FilterableVariant = IVariant & IFilterable;
