import { IAuth, IAuthMethod } from '~/types/Auth';

export interface IAuthMethodJWT extends IAuthMethod {
  method: 'METHOD_JWT';
  metadata: {
    authorize_url: string;
    callback_url: string;
  };
}

export interface IAuthMethodJWTMetadata {
  'io.flipt.auth.jwt.email'?: string;
  'io.flipt.auth.jwt.name'?: string;
  'io.flipt.auth.jwt.picture'?: string;
}

export interface IAuthJWTInternal extends IAuth {
  metadata: IAuthMethodJWTMetadata;
}
