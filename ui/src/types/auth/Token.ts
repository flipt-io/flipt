import { IAuth } from '~/types/Auth';

export interface IAuthMethodTokenMetadata {
  'io.flipt.auth.token.name': string;
  'io.flipt.auth.token.description': string;
}

export interface IAuthTokenInternal extends IAuth {
  metadata: IAuthMethodTokenMetadata;
}

export interface IAuthTokenInternalList {
  authentications: IAuthTokenInternal[];
}

export interface IAuthTokenBase {
  name: string;
  description: string;
  expiresAt?: string;
}

export interface IAuthToken extends IAuthTokenBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface IAuthTokenSecret {
  clientToken: string;
  authentication: IAuthTokenInternal;
}
