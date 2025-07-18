import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

import { ICommand } from '~/types/Cli';
import { ICurlOptions } from '~/types/Curl';
import { IEnvironmentConfiguration } from '~/types/Environment';

import { defaultHeaders } from '~/data/api';

export function cls(...args: ClassValue[]) {
  return twMerge(clsx(args));
}

export function getRevision(): string {
  const revision = localStorage.getItem('revision');
  if (!revision) {
    throw new Error('No revision found');
  }
  return revision;
}

export async function copyTextToClipboard(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }

  // Fallback: create a hidden <textarea>, select its contents, and
  // use the execCommand('copy').
  const textarea = document.createElement('textarea');
  textarea.value = text;

  // Avoid scrolling to the bottom
  textarea.style.position = 'fixed';
  textarea.style.top = '0';
  textarea.style.left = '0';
  textarea.style.opacity = '0';
  textarea.setAttribute('readonly', '');

  document.body.appendChild(textarea);

  textarea.select();
  textarea.setSelectionRange(0, text.length); // for iOS

  const successful = document.execCommand('copy');
  document.body.removeChild(textarea);

  if (!successful) {
    throw new Error('Fallback clipboard copy failed');
  }
}

export function upperFirst(word: string) {
  if (!word) return '';
  return word.charAt(0).toUpperCase() + word.slice(1);
}

export function titleCase(str: string) {
  if (!str) return '';
  return str.toLowerCase().split(' ').map(upperFirst).join(' ');
}

export function stringAsKey(str: string) {
  if (!str) return '';
  return str.toLowerCase().split(/\s+/).join('-');
}

const namespaces = '/namespaces/';
export function addNamespaceToPath(path: string, key: string): string {
  if (path.startsWith(namespaces)) {
    // [0] before slash ('')
    // [1] /namespaces/
    // [2] namespace key
    // [...] after slash
    const [, , existingKey, ...parts] = path.split('/');
    if (existingKey === key) {
      return path;
    }
    return `${namespaces}${key}/${parts.join('/')}`;
  }
  return `${namespaces}${key}${path}`;
}

type ErrorWithMessage = {
  message: string;
};

function isErrorWithMessage(error: unknown): error is ErrorWithMessage {
  return (
    typeof error === 'object' &&
    error !== null &&
    'message' in error &&
    typeof (error as Record<string, unknown>).message === 'string'
  );
}

function isFetchBaseQueryError(error: unknown): error is ErrorWithMessage {
  return (
    typeof error === 'object' &&
    error !== null &&
    'data' in error &&
    error.data !== null &&
    typeof (error.data as Record<string, unknown>).message === 'string'
  );
}

function toErrorWithMessage(maybeError: unknown): ErrorWithMessage {
  if (isErrorWithMessage(maybeError)) return maybeError;
  // handle Redux FetchBaseQueryError
  if (isFetchBaseQueryError(maybeError)) {
    // @ts-ignore
    return maybeError.data;
  }
  try {
    return new Error(JSON.stringify(maybeError));
  } catch {
    // fallback in case there's an error stringifying the maybeError
    // like with circular references for example.
    return new Error(String(maybeError));
  }
}

export function getErrorMessage(error: unknown) {
  return toErrorWithMessage(error).message;
}

export function generateCurlCommand(curlOptions: ICurlOptions) {
  const headers = { ...defaultHeaders(), ...curlOptions.headers };
  const curlHeaders = Object.keys(headers)
    .map((key) => `-H "${key}: ${headers[key]}"`)
    .join(' ');

  const curlData = `-d '${JSON.stringify(curlOptions.body)}'`;
  return [
    'curl',
    `-X ${curlOptions.method}`,
    curlHeaders,
    curlData,
    curlOptions.uri
  ].join(' ');
}

export function generateCliCommand(command: ICommand): string {
  return `flipt ${command.commandName} ${command.arguments?.join(' ')} ${command.options?.map(({ key, value }) => `${key} ${value}`).join(' ')}`;
}

export function getRepoUrlFromConfig(configuration: IEnvironmentConfiguration) {
  // TODO: support other SCMs as this is only works for GitHub/GitLab currently
  let repoUrl = configuration.remote;
  if (configuration.branch && configuration.directory) {
    repoUrl += `/tree/${configuration.branch}/${configuration.directory}`;
  } else if (configuration.branch) {
    repoUrl += `/tree/${configuration.branch}`;
  }
  return repoUrl;
}

export function extractRepoName(remote: string): string {
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
