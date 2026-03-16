import { faDiscord, faGithub } from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

export default function Footer() {
  const info = useSelector(selectInfo);

  const ref = () => {
    if (info?.isRelease && info?.version) {
      return info.version;
    }
    if (info?.commit) {
      return info.commit.substring(0, 7);
    }
    return 'v1-dev';
  };

  const refURL = () => {
    if (info?.isRelease && info?.version) {
      return `https://github.com/flipt-io/flipt/releases/tag/${info.version}`;
    }
    if (info?.commit) {
      return `https://github.com/flipt-io/flipt/commit/${info?.commit}`;
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
    <footer className="body-font text-secondary-foreground sticky top-[100vh] dark:text-gray-300">
      <div className="mt-4 flex flex-col items-center px-8 py-4 sm:flex-row">
        <div className="container mx-auto flex flex-col items-center space-x-4 sm:flex-row">
          <p className="text-muted-foreground dark:text-muted-foreground mt-4 text-xs sm:mt-0">
            <span className="hidden sm:inline">
              {ref() && (
                <>
                  <a href={refURL()} className="text-brand">
                    {ref()}
                  </a>
                  &nbsp;|&nbsp;
                </>
              )}
            </span>
            <span className="text-muted-foreground block sm:inline">
              &copy; {new Date().getFullYear()} Flipt Software Inc. All rights
              reserved.
            </span>
          </p>
        </div>
        <span className="mt-4 inline-flex justify-center space-x-5 sm:mt-0 sm:ml-auto sm:justify-start">
          {social.map((item) => (
            <a
              key={item.name}
              href={item.href}
              className="text-muted-foreground"
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
