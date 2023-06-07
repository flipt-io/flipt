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

export interface IStorage {
  type: StorageType;
}

// export interface IAuthentication {
//   required?: boolean;
// }

export interface IConfig {
  storage: IStorage;
  //authentication: IAuthentication;
}

export enum StorageType {
  DATABASE = 'database',
  GIT = 'git',
  LOCAL = 'local'
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
