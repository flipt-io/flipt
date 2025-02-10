import { IPageable } from './Pageable';

export interface INamespace {
  key: string;
  name: string;
  description?: string;
  protected: boolean;
}

export interface INamespaceList extends IPageable {
  items: INamespace[];
  revision?: string;
}
