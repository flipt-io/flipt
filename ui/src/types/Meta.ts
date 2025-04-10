import { Theme } from './Preferences';

export interface IInfo {
  build: IBuild;
  analytics?: IAnalytics;
  ui?: IUI;
}

export interface IBuild {
  version: string;
  latestVersion?: string;
  latestVersionURL?: string;
  commit: string;
  buildDate: string;
  updateAvailable: boolean;
  isRelease: boolean;
}

export interface IUI {
  defaultTheme: Theme;
  topbarColor?: string;
}

export interface IAnalytics {
  enabled: boolean;
}

export enum LoadingStatus {
  IDLE = 'idle',
  LOADING = 'loading',
  SUCCEEDED = 'succeeded',
  FAILED = 'failed'
}
