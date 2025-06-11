import { faDiscord, faGithub } from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

export default function Footer() {
  const info = useSelector(selectInfo);

  const ref = () => {
    if (info?.build?.isRelease && info?.build?.version) {
      return info.build.version;
    }
    if (info?.build?.commit) {
      return info.build.commit.substring(0, 7);
    }
    return 'Development';
  };

  const refURL = () => {
    if (info?.build?.isRelease && info?.build?.version) {
      return `https://github.com/flipt-io/flipt/releases/tag/${info.build.version}`;
    }
    if (info?.build?.commit) {
      return `https://github.com/flipt-io/flipt/commit/${info?.build?.commit}`;
    }
    return 'https://github.com/flipt-io/flipt';
  };

  const social = [
    {
      name: 'GitHub',
      href: 'https://www.github.com/flipt-io/flipt',
      icon: faGithub
    },
    {
      name: 'Discord',
      href: 'https://www.flipt.io/discord',
      icon: faDiscord
    }
  ];

  return (
    <footer className="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300">
      <div className="container mx-auto flex flex-col sm:flex-row items-center justify-between px-8 py-4 gap-4">
        {/* Version + Badge */}
        <div className="flex items-center space-x-2">
          {ref() && (
            <a
              href={refURL()}
              className="text-violet-500 dark:text-violet-400 font-medium text-xs hover:underline focus:outline-none focus:ring-2 focus:ring-violet-400 rounded"
              title="View build details on GitHub"
              aria-label="View build details on GitHub"
            >
              {ref()}
            </a>
          )}
          {info?.product && (
            <a
              href="https://docs.flipt.io/v2/editions"
              target="_blank"
              rel="noopener noreferrer"
              className={`inline-block rounded-full px-2 py-0.5 text-xs font-semibold align-middle transition-colors duration-150
                ${
                  info.product === 'enterprise'
                    ? 'bg-violet-500 text-white dark:bg-violet-400 dark:text-gray-900 hover:bg-violet-600 dark:hover:bg-violet-500'
                    : 'border border-violet-500 text-violet-500 bg-transparent hover:bg-violet-50 dark:hover:bg-violet-900'
                }
              `}
              title={
                info.product === 'enterprise'
                  ? 'Enterprise Edition'
                  : 'Open Source Edition'
              }
              aria-label={
                info.product === 'enterprise'
                  ? 'Enterprise Edition'
                  : 'Open Source Edition'
              }
            >
              {info.product === 'enterprise' ? 'Enterprise' : 'Open Source'}
            </a>
          )}
        </div>
        {/* Copyright */}
        <div className="text-xs text-gray-500 dark:text-gray-400 text-center sm:text-left">
          &copy; {new Date().getFullYear()} Flipt Software Inc. All rights
          reserved.
        </div>
        {/* Social */}
        <div className="flex items-center space-x-4">
          {social.map((item) => (
            <a
              key={item.name}
              href={item.href}
              className="text-gray-400 hover:text-violet-500 dark:hover:text-violet-400 transition-colors duration-150 p-1 rounded-full focus:outline-none focus:ring-2 focus:ring-violet-400"
              title={item.name}
              aria-label={item.name}
            >
              <FontAwesomeIcon
                icon={item.icon}
                className="h-5 w-5"
                aria-hidden={true}
              />
            </a>
          ))}
        </div>
      </div>
    </footer>
  );
}
