import { useSelector } from 'react-redux';
import { selectConfig } from '~/app/meta/metaSlice';
import { titleCase } from '~/utils/helpers';
import { IconDefinition, faGitAlt } from '@fortawesome/free-brands-svg-icons';
import {
  faCube,
  faCloud,
  faDatabase,
  faFileCode
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const storageTypes: Record<string, IconDefinition> = {
  local: faFileCode,
  object: faCloud,
  git: faGitAlt,
  database: faDatabase,
  oci: faCube
};

export default function ReadOnly() {
  const config = useSelector(selectConfig);

  const storageIcon = config.storage?.type
    ? storageTypes[config.storage?.type]
    : undefined;

  return (
    <span
      className="nightwind-prevent text-white inline-flex items-center gap-x-1.5 rounded-md px-3 py-1 text-xs font-medium"
      title={`Backed by ${titleCase(
        config.storage?.type || 'unknown'
      )} Storage`}
    >
      {config.storage?.type == 'git' && (
        <span
          className="nightwind-prevent text-violet-200 max-w-32 truncate font-medium"
          title={`ref: ${config.storage?.git?.ref} repo: ${config.storage?.git?.repository}`}
        >
          {config.storage?.git?.ref}
        </span>
      )}
      {storageIcon && (
        <FontAwesomeIcon
          icon={storageIcon}
          className="text-gray h-5 w-5"
          aria-hidden={true}
        />
      )}
      Read-Only
    </span>
  );
}
