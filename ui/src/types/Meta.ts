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
  readOnly?: boolean;
}

export interface IConfig {
  storage: IStorage;
}

export enum StorageType {
  DATABASE = 'database',
  GIT = 'git',
  LOCAL = 'local',
  OBJECT = 'object'
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
