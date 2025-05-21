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
  base?: string;
}

export interface IBranchEnvironment {
  baseEnvironmentKey: string;
  environmentKey: string;
  branch: string;
}
