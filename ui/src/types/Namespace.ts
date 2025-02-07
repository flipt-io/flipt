import { IPageable } from './Pageable';

export interface INamespaceBase {
  key: string;
  name: string;
  description?: string;
}

export interface INamespace extends INamespaceBase {
  protected: boolean;
}

export interface INamespaceList extends IPageable {
  items: INamespace[];
}
