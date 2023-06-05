export interface IInfo {
  version: string;
  latestVersion?: string;
  latestVersionURL?: string;
  commit: string;
  buildDate: string;
  goVersion: string;
  updateAvailable: boolean;
  isRelease: boolean;
}

export interface IDb {
  url: string;
}

export interface IAuthentication {
  required?: boolean;
}

export interface IConfig {
  db: IDb;
  authentication: IAuthentication;
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
