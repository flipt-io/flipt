import FlagForm from '~/components/flags/forms/FlagForm';
import MoreInfo from '~/components/MoreInfo';
import { PageHeader } from '~/components/ui/page';

export default function NewFlag() {
  return (
    <>
      <PageHeader title="New Flag" />
      <div className="mb-8 space-y-4">
        <MoreInfo href="https://www.flipt.io/docs/concepts#flags">
          Learn more about flags
        </MoreInfo>
      </div>

      <div className="mb-8 max-w-screen-md">
        <FlagForm />
      </div>
    </>
  );
}
