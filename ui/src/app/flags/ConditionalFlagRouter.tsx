import { useOutletContext } from 'react-router-dom';
import { FlagType, IFlag } from '~/types/Flag';
import Variants from '~/components/variants/Variants';
import Rollouts from '~/components/rollouts/Rollouts';

type ConditionalFlagRouterProps = {
  flag: IFlag;
};

export default function ConditionalFlagRouter() {
  const { flag } = useOutletContext<ConditionalFlagRouterProps>();

  return (
    <>
      {flag.type === FlagType.VARIANT ? (
        <Variants flag={flag} />
      ) : (
        <Rollouts flag={flag} />
      )}
    </>
  );
}
