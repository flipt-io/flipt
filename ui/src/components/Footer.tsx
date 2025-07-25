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
    <footer className="body-font sticky top-[100vh] text-gray-700 dark:text-gray-300">
      <div className="mt-4 flex flex-col items-center px-8 py-4 sm:flex-row">
        <div className="container mx-auto flex flex-col items-center space-x-4 sm:flex-row">
          <p className="mt-4 text-xs text-gray-500 dark:text-gray-400 sm:mt-0">
            <span className="hidden sm:inline">
              {ref() && (
                <>
                  <a
                    href={refURL()}
                    className="text-violet-500 dark:text-violet-400"
                  >
                    {ref()}
                  </a>
                  &nbsp;|&nbsp;
                </>
              )}
            </span>
            <span className="block sm:inline">
              &copy; {new Date().getFullYear()} Flipt Software Inc. All rights
              reserved.
            </span>
          </p>
        </div>
        <span className="mt-4 inline-flex justify-center space-x-5 sm:ml-auto sm:mt-0 sm:justify-start">
          {social.map((item) => (
            <a
              key={item.name}
              href={item.href}
              className="text-muted-foreground hover:text-gray-500 dark:hover:text-gray-100"
            >
              <span className="sr-only">{item.name}</span>
              <FontAwesomeIcon
                icon={item.icon}
                className="text-gray h-5 w-5"
                aria-hidden={true}
              />
            </a>
          ))}
        </span>
      </div>
    </footer>
  );
}
