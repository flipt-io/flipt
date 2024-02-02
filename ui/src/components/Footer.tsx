import {
  faDiscord,
  faGithub,
  faXTwitter
} from '@fortawesome/free-brands-svg-icons';
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
    return '';
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
      name: 'Twitter',
      href: 'https://www.twitter.com/flipt_io',
      icon: faXTwitter
    },
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
    <footer className="body-font text-gray-700 sticky top-[100vh]">
      <div className="mt-4 flex flex-col items-center px-8 py-4 sm:flex-row">
        <div className="container mx-auto flex flex-col items-center space-x-4 sm:flex-row">
          <p className="text-gray-500 mt-4 text-xs sm:mt-0">
            <span className="hidden sm:inline">
              {ref() && (
                <>
                  <a href={refURL()} className="text-violet-500">
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
          <p className="text-gray-500 mt-4 text-xs sm:mt-0">
            <span className="hidden sm:inline">
              <a
                target="_blank"
                rel="noreferrer"
                href="https://features.flipt.io/changelog"
                className="text-violet-500"
              >
                Changelog
              </a>
            </span>
          </p>
          <p className="text-gray-500 mt-4 text-xs sm:mt-0">
            <span className="hidden sm:inline">
              <a
                target="_blank"
                rel="noreferrer"
                href="https://features.flipt.io"
                className="text-violet-500"
              >
                Share Feedback
              </a>
            </span>
          </p>
        </div>
        <span className="mt-4 inline-flex justify-center space-x-5 sm:ml-auto sm:mt-0 sm:justify-start">
          {social.map((item) => (
            <a
              key={item.name}
              href={item.href}
              className="text-gray-400 hover:text-gray-500"
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
