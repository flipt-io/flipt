import { IAuth, IAuthMethod } from '~/types/Auth';

export interface IAuthMethodGithub extends IAuthMethod {
  method: 'METHOD_GITHUB';
  metadata: {
    authorize_url: string;
    callback_url: string;
  };
}

export interface IAuthMethodGithubMetadata {
  'io.flipt.auth.oauth.email'?: string;
  'io.flipt.auth.oauth.user'?: string;
}

export interface IAuthOIDCInternal extends IAuth {
  metadata: IAuthMethodGithubMetadata;
}
