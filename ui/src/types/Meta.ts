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
  analyticsEnabled: boolean;
}

export interface IUI {
  defaultTheme: Theme;
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
