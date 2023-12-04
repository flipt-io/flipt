import {
  CircleStackIcon,
  CloudIcon,
  CodeBracketIcon,
  CubeIcon,
  DocumentIcon
} from '@heroicons/react/20/solid';
import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';
import { Icon } from '~/types/Icon';
import { titleCase } from '~/utils/helpers';

const storageTypes: Record<string, Icon> = {
  local: DocumentIcon,
  object: CloudIcon,
  git: CodeBracketIcon,
  database: CircleStackIcon,
  oci: CubeIcon
};

export default function ReadOnly() {
  const config = useSelector(selectConfig);

  const StorageIcon = config.storage?.type
    ? storageTypes[config.storage?.type]
    : undefined;

  return (
    <span
      className="nightwind-prevent text-white inline-flex items-center gap-x-1.5 rounded-md px-3 py-1 text-xs font-medium"
      title={`Backed by ${titleCase(
        config.storage?.type || 'unknown'
      )} Storage`}
    >
      {StorageIcon && (
        <StorageIcon className="h-4 w-4 fill-gray-200" aria-hidden="true" />
      )}
      Read-Only
    </span>
  );
}
