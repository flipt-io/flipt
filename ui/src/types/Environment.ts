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

export enum SCM {
  GITHUB = 'GITHUB_SCM'
}

export enum ProposalState {
  OPEN = 'PROPOSAL_STATE_OPEN',
  MERGED = 'PROPOSAL_STATE_MERGED'
  //TODO: CLOSED = "PROPOSAL_STATE_CLOSED",
}

export interface IEnvironmentProposal {
  scm: SCM;
  url: string;
  state?: ProposalState;
}
