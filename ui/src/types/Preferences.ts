export enum TimezoneType {
  UTC = 'utc',
  LOCAL = 'local'
}

export type Preferences = {
  timezone: TimezoneType;
};
