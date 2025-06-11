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
  GITHUB = 'GITHUB_SCM',
  GITLAB = 'GITLAB_SCM',
  GITEA = 'GITEA_SCM'
}

export enum ProposalState {
  OPEN = 'PROPOSAL_STATE_OPEN',
  MERGED = 'PROPOSAL_STATE_MERGED',
  CLOSED = 'PROPOSAL_STATE_CLOSED'
}

export interface IEnvironmentProposal {
  url: string;
  state?: ProposalState;
}
