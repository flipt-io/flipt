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
  scm?: SCM;
}

export interface IBranchEnvironment {
  environmentKey: string;
  key: string;
  branch: string;
  proposal?: IEnvironmentProposal;
}

export enum SCM {
  UNKNOWN = 'SCM_UNKNOWN',
  GITHUB = 'SCM_GITHUB',
  GITLAB = 'SCM_GITLAB',
  GITEA = 'SCM_GITEA',
  AZURE = 'SCM_AZURE',
  BITBUCKET = 'SCM_BITBUCKET'
}

export enum ProposalState {
  UNKNOWN = 'PROPOSAL_STATE_UNKNOWN',
  OPEN = 'PROPOSAL_STATE_OPEN',
  MERGED = 'PROPOSAL_STATE_MERGED',
  CLOSED = 'PROPOSAL_STATE_CLOSED'
}

export interface IEnvironmentProposal {
  url: string;
  state?: ProposalState;
}
