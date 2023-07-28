import { ChevronRightIcon } from '@heroicons/react/20/solid';
import {
  Bars4Icon,
  BookOpenIcon,
  RssIcon,
  UserGroupIcon
} from '@heroicons/react/24/outline';
import { Link } from 'react-router-dom';
import logoFlag from '~/assets/logo-flag.png';

const links = [
  {
    title: 'Documentation',
    description: 'Learn how to integrate Flipt with your app',
    href: 'https://www.flipt.io/docs',
    icon: BookOpenIcon
  },
  {
    title: 'API Reference',
    description: 'A complete API reference for our REST API',
    href: 'https://www.flipt.io/docs/reference/overview',
    icon: Bars4Icon
  },
  {
    title: 'Community',
    description:
      'Join our Discord community to chat with other users and contributors',
    href: 'https://www.flipt.io/discord',
    icon: UserGroupIcon
  },
  {
    title: 'Blog',
    description: 'Read our latest news and articles',
    href: 'https://www.flipt.io/blog',
    icon: RssIcon
  }
];

export default function NotFoundLayout() {
  return (
    <div className="bg-white flex min-h-screen flex-col">
      <main className="mx-auto w-full max-w-7xl px-6 lg:px-8">
        <div className="flex-shrink-0 pt-16">
          <Link to="/">
            <img
              src={logoFlag}
              alt="logo"
              width={512}
              height={512}
              className="m-auto h-20 w-auto"
            />
          </Link>
        </div>
        <div className="mx-auto max-w-xl py-16 sm:py-24">
          <div className="text-center">
            <p className="text-violet-600 text-base font-semibold">404</p>
            <h1 className="text-gray-900 mt-2 text-4xl font-bold tracking-tight sm:text-5xl">
              Not Found
            </h1>
            <p className="text-gray-500 mt-2 text-lg">
              The page you are looking for could not be found.
            </p>
          </div>
          <div className="mt-12">
            <h2 className="text-gray-500 text-base font-semibold">
              Popular pages
            </h2>
            <ul
              role="list"
              className="border-gray-200 mt-4 divide-y divide-gray-200 border-b border-t"
            >
              {links.map((link, linkIdx) => (
                <li
                  key={linkIdx}
                  className="relative flex items-start space-x-4 py-6"
                >
                  <div className="flex-shrink-0">
                    <span className="bg-violet-50 flex h-12 w-12 items-center justify-center rounded-lg">
                      <link.icon
                        className="text-violet-700 h-6 w-6"
                        aria-hidden="true"
                      />
                    </span>
                  </div>
                  <div className="min-w-0 flex-1">
                    <h3 className="text-gray-900 text-base font-medium hover:text-violet-500">
                      <span className="rounded-sm focus-within:ring-2 focus-within:ring-violet-500 focus-within:ring-offset-2">
                        <a
                          href={link.href}
                          className="focus:outline-none"
                          target="_blank"
                          rel="noreferrer"
                        >
                          <span
                            className="absolute inset-0"
                            aria-hidden="true"
                          />
                          {link.title}
                        </a>
                      </span>
                    </h3>
                    <p className="text-gray-500 text-base">
                      {link.description}
                    </p>
                  </div>
                  <div className="flex-shrink-0 self-center">
                    <ChevronRightIcon
                      className="text-gray-400 h-5 w-5"
                      aria-hidden="true"
                    />
                  </div>
                </li>
              ))}
            </ul>
            <div className="mt-8">
              <a
                href="/"
                className="text-violet-600 text-base font-medium hover:text-violet-500"
              >
                <span aria-hidden="true">&larr; </span>
                Home
              </a>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
