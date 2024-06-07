import { IAuth, IAuthMethod } from '~/types/Auth';

export interface IAuthMethodGithub extends IAuthMethod {
  method: 'METHOD_GITHUB';
  metadata: {
    authorize_url: string;
  };
}

export interface IAuthMethodGithubMetadata {
  'io.flipt.auth.github.email'?: string;
  'io.flipt.auth.github.name'?: string;
  'io.flipt.auth.github.picture'?: string;
  'io.flipt.auth.github.preferred_username'?: string;
}

export interface IAuthGithubInternal extends IAuth {
  method: 'METHOD_GITHUB';
  metadata: IAuthMethodGithubMetadata;
}
