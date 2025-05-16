export interface IEnvironment {
  key: string;
  name?: string;
  default?: boolean;
  configuration?: IEnvironmentConfiguration;
}

export interface IEnvironmentConfiguration {
  remote: string;
  branch: string;
  directory: string;
}
