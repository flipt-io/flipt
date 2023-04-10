import { IFlag } from '~/types/Flag';

export type FlagProps = {
  flag: IFlag;
  onFlagChange: () => void;
};
