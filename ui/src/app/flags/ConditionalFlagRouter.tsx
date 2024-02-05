import { useOutletContext } from 'react-router-dom';
import { FlagType, IFlag } from '~/types/Flag';
import Rollouts from './rollouts/Rollouts';
import Variants from './variants/Variants';

type ConditionalFlagRouterProps = {
  flag: IFlag;
};

export default function ConditionalFlagRouter() {
  const { flag } = useOutletContext<ConditionalFlagRouterProps>();

  return (
    <>
      {flag.type === FlagType.VARIANT ? (
        <>
          <Variants flag={flag} />
        </>
      ) : (
        <>
          <Rollouts flag={flag} />
        </>
      )}
    </>
  );
}
