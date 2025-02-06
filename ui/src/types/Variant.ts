import { IFilterable } from './Selectable';

export interface IVariant {
  key: string;
  name: string;
  description: string;
  attachment?: Record<string, any>;
}

export type FilterableVariant = Pick<IVariant, 'key' | 'name'> & IFilterable;

export function toFilterableVariant(selected: IVariant | null) {
  if (selected) {
    return {
      ...selected,
      displayValue: selected.name,
      filterValue: selected.key
    };
  }
  return null;
}
