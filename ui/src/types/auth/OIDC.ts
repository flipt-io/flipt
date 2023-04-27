import { IAuth, IAuthMethod } from '~/types/Auth';

interface AuthMethodOIDCMetadataProvider {
  authorize_url: string;
  callback_url: string;
}

export interface IAuthMethodOIDC extends IAuthMethod {
  method: 'METHOD_OIDC';
  metadata: {
    providers: Record<string, AuthMethodOIDCMetadataProvider>;
  };
}

export interface IAuthMethodOIDCMetadata {
  'io.flipt.auth.oidc.email'?: string;
  'io.flipt.auth.oidc.email_verified'?: string;
  'io.flipt.auth.oidc.name'?: string;
  'io.flipt.auth.oidc.picture'?: string;
  'io.flipt.auth.oidc.provider': string;
}

export interface IAuthOIDCInternal extends IAuth {
  metadata: IAuthMethodOIDCMetadata;
}
