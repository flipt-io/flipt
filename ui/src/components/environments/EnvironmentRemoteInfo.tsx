import { FolderGit, Github, Gitlab } from 'lucide-react';

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
  const { configuration, key } = environment || {};
  if (!configuration?.remote) return null;

  // Determine provider icon
  let ProviderIcon = FolderGit;
  if (configuration.remote.includes('github.com')) ProviderIcon = Github;
  if (configuration.remote.includes('gitlab.com')) ProviderIcon = Gitlab;

  // Build repo link (optionally to directory)
  let repoUrl = configuration.remote;
  if (configuration.branch && configuration.directory) {
    repoUrl += `/tree/${configuration.branch}/${configuration.directory}`;
  } else if (configuration.branch) {
    repoUrl += `/tree/${configuration.branch}`;
  }

  // Use improved extraction for repo name
  const repoName = extractRepoName(configuration.remote);

  return (
    <div className="mx-2 my-3 rounded-lg bg-white/80 dark:bg-muted/60 shadow-sm border border-muted flex flex-col gap-1 p-3">
      <div className="flex items-center gap-2 mb-1">
        <ProviderIcon className="w-4 h-4 text-muted-foreground" />
        <a
          href={repoUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="font-medium text-sm text-foreground hover:underline break-all truncate max-w-[140px]"
          title={repoName}
        >
          {repoName}
        </a>
      </div>
      <div className="flex gap-2 mt-1">
        {configuration.branch && (
          <span
            className="font-mono text-xs bg-muted rounded px-1 py-0.5 truncate max-w-[60px]"
            title={configuration.branch}
          >
            {configuration.branch}
          </span>
        )}
        {configuration.directory && (
          <span
            className="font-mono text-xs text-muted-foreground truncate max-w-[60px]"
            title={configuration.directory}
          >
            {configuration.directory}
          </span>
        )}
      </div>
    </div>
  );
}
