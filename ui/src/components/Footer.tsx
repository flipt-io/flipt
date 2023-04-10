import {
  faDiscord,
  faGithub,
  faMastodon,
  faTwitter
} from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useEffect, useState } from 'react';
import { getInfo } from '~/data/api';
import { Info } from '~/types/Meta';

export default function Footer() {
  const [info, setInfo] = useState<Info | null>(null);

  useEffect(() => {
    getInfo()
      .then((info: Info) => {
        setInfo(info);
      })
      .catch(() => {
        // nothing to do, component will degrade gracefully
      });
  }, []);

  const ref = () => {
    if (info?.isRelease && info?.version) {
      return `v${info.version}`;
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
      icon: faTwitter
    },
    {
      name: 'Mastadon',
      href: 'https://www.hachyderm.io/@flipt',
      icon: faMastodon
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
    <footer className="body-font sticky top-[100vh] text-gray-700">
      <div className="container mx-auto mt-4 flex flex-col items-center px-8 py-4 sm:flex-row">
        <p className="mt-4 text-xs text-gray-500 sm:mt-0">
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
