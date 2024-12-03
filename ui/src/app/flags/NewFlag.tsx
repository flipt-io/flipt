import FlagForm from '~/components/flags/forms/FlagForm';
import MoreInfo from '~/components/MoreInfo';
import { PageHeader } from '~/components/ui/page';

export default function NewFlag() {
  return (
    <>
      <PageHeader title="New Flag" />
      <div className="my-6">
        <div className="md:grid md:grid-cols-3 md:gap-6">
          <div className="md:col-span-1">
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
