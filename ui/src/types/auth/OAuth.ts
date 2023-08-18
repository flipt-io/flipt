import { IAuth, IAuthMethod } from '~/types/Auth';

interface AuthMethodOAuthMetadataProvider {
  authorize_url: string;
  callback_url: string;
}

export interface IAuthMethodOAuth extends IAuthMethod {
  method: 'METHOD_OAUTH';
  metadata: {
    hosts: Record<string, AuthMethodOAuthMetadataProvider>;
  };
}

export interface IAuthMethodOAuthMetadata {
  'io.flipt.auth.oauth.access_token'?: string;
}

export interface IAuthOIDCInternal extends IAuth {
  metadata: IAuthMethodOAuthMetadata;
}
