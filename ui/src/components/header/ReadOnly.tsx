import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';
import { titleCase } from '~/utils/helpers';
import { Box, Database, FolderGit2, Cloud, FileCode } from 'lucide-react';

const storageTypes: Record<string, any> = {
  local: FileCode,
  object: Cloud,
  git: FolderGit2,
  database: Database,
  oci: Box
};

export default function ReadOnly() {
  const config = useSelector(selectConfig);

  const StorageIcon = config.storage?.type
    ? storageTypes[config.storage?.type]
    : undefined;

  return (
    <span
      className="inline-flex items-center gap-x-1.5 rounded-md p-1 text-xs font-medium text-white"
      title={`Backed by ${titleCase(
        config.storage?.type || 'unknown'
      )} Storage`}
    >
      {config.storage?.type == 'git' && (
        <span
          className="max-w-32 truncate font-medium text-violet-200"
          title={`ref: ${config.storage?.git?.ref} repo: ${config.storage?.git?.repository}`}
        >
          {config.storage?.git?.ref}
        </span>
      )}
      {StorageIcon && (
        <StorageIcon className="text-foreground h-4 w-4" aria-hidden={true} />
      )}
      Read-Only
    </span>
  );
}
