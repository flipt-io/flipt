import { IAuth, IAuthMethod } from '~/types/Auth';

export interface IAuthMethodJWT extends IAuthMethod {
  method: 'METHOD_JWT';
  metadata: {};
}

export interface IAuthMethodJWTMetadata {
  'io.flipt.auth.jwt.email'?: string;
  'io.flipt.auth.jwt.name'?: string;
  'io.flipt.auth.jwt.picture'?: string;
  'io.flipt.auth.jwt.issuer'?: string;
}

export interface IAuthJWTInternal extends IAuth {
  method: 'METHOD_JWT';
  metadata: IAuthMethodJWTMetadata;
}
