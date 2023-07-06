import { useOutletContext } from 'react-router-dom';
import FlagForm from '~/components/flags/FlagForm';
import MoreInfo from '~/components/MoreInfo';
import { FlagType } from '~/types/Flag';
import { FlagProps } from './FlagProps';
import Rollouts from './rollouts/Rollouts';
import Variants from './variants/Variants';

export default function EditFlag() {
  const { flag, onFlagChange } = useOutletContext<FlagProps>();

  const flagTypeToLabel = (t: string) => FlagType[t as keyof typeof FlagType];

  return (
    <>
      <div className="flex flex-col">
        {/* flag details */}
        <div className="my-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <p className="mt-1 text-sm text-gray-500">
                Basic information about the flag and its status.
              </p>
              <MoreInfo
                className="mt-5"
                href="https://www.flipt.io/docs/concepts#flags"
              >
                Learn more about flags
              </MoreInfo>
            </div>
            <div className="mt-5 md:col-span-2 md:mt-0">
              <FlagForm flag={flag} flagChanged={onFlagChange} />
            </div>
          </div>
        </div>

        {flagTypeToLabel(flag.type) === FlagType.VARIANT_FLAG_TYPE && (
          <Variants flag={flag} flagChanged={onFlagChange} />
        )}
        {flagTypeToLabel(flag.type) === FlagType.BOOLEAN_FLAG_TYPE && (
          <Rollouts flag={flag} flagChanged={onFlagChange} />
        )}
      </div>
    </>
  );
}
