import { useOutletContext } from 'react-router-dom';
import { FlagType, IFlag } from '~/types/Flag';
import VariantFlag from '~/components/flags/VariantFlag';
import BooleanFlag from '~/components/flags/BooleanFlag';

type ConditionalFlagRouterProps = {
  flag: IFlag;
};

export default function ConditionalFlagRouter() {
  const { flag } = useOutletContext<ConditionalFlagRouterProps>();

  return (
    <>
      {flag.type === FlagType.VARIANT ? (
        <VariantFlag flag={flag} />
      ) : (
        <BooleanFlag flag={flag} />
      )}
    </>
  );
}
