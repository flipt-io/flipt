import MoreInfo from '~/components/MoreInfo';
import { PageHeader } from '~/components/Page';
import FlagForm from '~/components/flags/FlagForm';

export default function NewFlag() {
  return (
    <>
      <PageHeader title="New Flag" />
      <div className="mb-8 space-y-4">
        <MoreInfo href="https://docs.flipt.io/v2/concepts#flags">
          Learn more about flags
        </MoreInfo>
      </div>

      <div className="mb-8">
        <FlagForm />
      </div>
    </>
  );
}
