import { IFilterable } from './Selectable';

export interface IVariantBase {
  key: string;
  name: string;
  description: string;
  attachment?: Record<string, any>;
}

export interface IVariant extends IVariantBase {
  id: string;
}

export type FilterableVariant = Pick<IVariant, 'id' | 'key' | 'name'> &
  IFilterable;

export function toFilterableVariant(selected: IVariant | null) {
  if (selected) {
    return {
      ...selected,
      displayValue: selected.name,
      filterValue: selected.id
    };
  }
  return null;
}
