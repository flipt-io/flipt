import { FolderGit, Github, Gitlab } from 'lucide-react';

import { Badge } from '~/components/Badge';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger
} from '~/components/ui/tooltip';

import { IEnvironment } from '~/types/Environment';

function extractRepoName(remote: string): string {
  if (!remote) return '';
  // Remove protocol and trailing .git
  let url = remote.replace(/^https?:\/\//, '').replace(/\.git$/, '');

  // Handle SSH URLs
  if (url.includes('@')) {
    // git@github.com:org/repo.git or git@gitlab.com:group/subgroup/repo.git
    url = url.split(':')[1] || '';
    url = url.replace(/\.git$/, '');
    return url;
  }

  // For HTTP(S) URLs, remove domain
  const parts = url.split('/');
  // Remove the domain part (e.g., github.com)
  parts.shift();
  return parts.join('/');
}

export function EnvironmentRemoteInfo({
  environment
}: {
  environment: IEnvironment;
}) {
  const { configuration } = environment || {};
  if (!configuration?.remote) return null;

  let ProviderIcon = FolderGit;
  if (configuration.remote.includes('github.com')) ProviderIcon = Github;
  if (configuration.remote.includes('gitlab.com')) ProviderIcon = Gitlab;

  let repoUrl = configuration.remote;
  if (configuration.branch && configuration.directory) {
    repoUrl += `/tree/${configuration.branch}/${configuration.directory}`;
  } else if (configuration.branch) {
    repoUrl += `/tree/${configuration.branch}`;
  }

  const repoName = extractRepoName(configuration.remote);

  return (
    <a
      href={repoUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center gap-1 text-inherit hover:underline"
      title={repoUrl}
      style={{ textDecoration: 'none' }}
    >
      <Tooltip>
        <TooltipTrigger asChild>
          <Badge
            variant="secondary"
            className="flex items-center gap-1 px-2 py-1 bg-background font-semibold text-xs"
          >
            <ProviderIcon className="w-4 h-4 text-muted-foreground" />
            <span className="truncate max-w-[200px]">{repoName}</span>
          </Badge>
        </TooltipTrigger>
        <TooltipContent>{`Branch: ${configuration.branch}`}</TooltipContent>
      </Tooltip>
    </a>
  );
}
