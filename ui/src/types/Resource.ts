import { IPageable } from './Pageable';

export interface IResourceResponse<T> {
  resource: IResource<T>;
  revision: string;
}

export interface IResource<T> {
  namespace: string;
  key: string;
  payload: T;
}

export interface IResourceListResponse<T> extends IPageable {
  resources: IResource<T>[];
  revision: string;
}
