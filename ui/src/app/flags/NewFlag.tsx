import FlagForm from '~/components/flags/FlagForm';
import MoreInfo from '~/components/MoreInfo';

export default function NewFlag() {
  return (
    <>
      <div className="lg:flex lg:items-center lg:justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-2xl font-semibold leading-6 text-gray-900">
            New Flag
          </h2>
        </div>
      </div>
      <div className="my-10">
        <div className="md:grid md:grid-cols-3 md:gap-6">
          <div className="md:col-span-1">
            <h3 className="text-lg font-medium leading-6 text-gray-900">
              Details
            </h3>
            <p className="mt-2 text-sm text-gray-500">
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
            <FlagForm />
          </div>
        </div>
      </div>
    </>
  );
}
