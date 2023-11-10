import {
  CircleStackIcon,
  CloudIcon,
  CodeBracketIcon,
  DocumentIcon
} from '@heroicons/react/20/solid';
import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';
import { Icon } from '~/types/Icon';

const storageTypes: Record<string, Icon> = {
  local: DocumentIcon,
  object: CloudIcon,
  git: CodeBracketIcon,
  database: CircleStackIcon
};

export default function ReadOnly() {
  const config = useSelector(selectConfig);

  const StorageIcon = config.storage?.type
    ? storageTypes[config.storage?.type]
    : undefined;

  return (
    <span
      className="nightwind-prevent text-white inline-flex items-center gap-x-1.5 rounded-md px-3 py-1 text-xs font-medium"
      title={`Backed by ${config.storage?.type || 'unknown'} storage`}
    >
      {StorageIcon && (
        <StorageIcon className="h-3 w-3 fill-gray-200" aria-hidden="true" />
      )}
      Read-Only
    </span>
  );
}
