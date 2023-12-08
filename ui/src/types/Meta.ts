import { Theme } from './Preferences';

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

export interface IUI {
  defaultTheme: Theme;
}

export interface IConfig {
  status: LoadingStatus;
  storage: IStorage;
  ui: IUI;
}

export enum StorageType {
  DATABASE = 'database',
  GIT = 'git',
  LOCAL = 'local',
  OBJECT = 'object',
  OCI = 'oci'
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
