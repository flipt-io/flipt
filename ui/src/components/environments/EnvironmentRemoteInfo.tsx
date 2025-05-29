import { FolderGit, Github, Gitlab } from 'lucide-react';

import { Badge } from '~/components/Badge';

import { IEnvironment } from '~/types/Environment';

import { extractRepoName, getRepoUrlFromConfig } from '~/utils/helpers';

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

  const repoUrl = getRepoUrlFromConfig(configuration);
  const repoName = extractRepoName(configuration.remote);

  return (
    <a
      href={repoUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="flex gap-1 text-inherit hover:underline ml-auto"
      style={{ textDecoration: 'none' }}
    >
      <Badge
        variant="secondary"
        className="flex items-center gap-2 px-2 py-1 bg-background font-semibold text-xs"
      >
        <ProviderIcon className="w-4 h-4 text-muted-foreground" />
        <span className="truncate max-w-[200px]">{repoName}</span>
      </Badge>
    </a>
  );
}
