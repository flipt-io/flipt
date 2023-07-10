export interface ISelectable {
  key: string;
  displayValue: string;
}

export interface IFilterable extends ISelectable {
  status?: 'active' | 'inactive';
  filterValue: string;
}
