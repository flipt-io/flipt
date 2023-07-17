import { useOutletContext } from 'react-router-dom';
import FlagForm from '~/components/flags/FlagForm';
import MoreInfo from '~/components/MoreInfo';
import { FlagType } from '~/types/Flag';
import { FlagProps } from './FlagProps';
import Rollouts from './rollouts/Rollouts';
import Variants from './variants/Variants';

export default function EditFlag() {
  const { flag, onFlagChange } = useOutletContext<FlagProps>();

  return (
    <>
      <div className="flex flex-col">
        {/* flag details */}
        <div className="my-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <p className="text-gray-500 mt-1 text-sm">
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

        {flag.type === FlagType.VARIANT && (
          <Variants flag={flag} flagChanged={onFlagChange} />
        )}
        {flag.type === FlagType.BOOLEAN && <Rollouts flag={flag} />}
      </div>
    </>
  );
}
