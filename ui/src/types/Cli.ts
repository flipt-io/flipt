export enum Command {
  Evaluate = 'evaluate'
}

export interface IOption {
  key: string;
  value: string;
}

export interface ICommand {
  commandName: Command;
  arguments?: string[];
  options?: IOption[];
}
