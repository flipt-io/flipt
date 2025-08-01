import {
  BookOpenIcon,
  ChevronRightIcon,
  RssIcon,
  UsersIcon
} from 'lucide-react';
import { Link } from 'react-router';

import logoFlag from '~/assets/logo-flag.png';

const links = [
  {
    title: 'Documentation',
    description: 'Learn how to integrate Flipt with your app',
    href: 'https://docs.flipt.io/v2/',
    icon: BookOpenIcon
  },
  {
    title: 'Community',
    description:
      'Join our Discord community to chat with other users and contributors',
    href: 'https://www.flipt.io/discord',
    icon: UsersIcon
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
    <div className="flex min-h-screen flex-col bg-sidebar">
      <main className="mx-auto w-full max-w-7xl px-6 lg:px-8">
        <div className="shrink-0 pt-16">
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
            <h1 className="mt-2 text-4xl font-bold tracking-tight text-secondary-foreground sm:text-5xl">
              Not Found
            </h1>
            <p className="mt-2 text-lg text-muted-foreground">
              The page you are looking for could not be found.
            </p>
          </div>
          <div className="mt-12">
            <h2 className="text-base font-semibold text-muted-foreground">
              Popular pages
            </h2>
            <ul role="list" className="mt-4 divide-y border-b border-t">
              {links.map((link, linkIdx) => (
                <li
                  key={linkIdx}
                  className="relative flex items-start space-x-4 py-6"
                >
                  <div className="shrink-0">
                    <span className="flex h-12 w-12 items-center justify-center rounded-lg">
                      <link.icon
                        className="h-6 w-6 text-brand"
                        aria-hidden="true"
                      />
                    </span>
                  </div>
                  <div className="min-w-0 flex-1">
                    <h3 className="text-base font-medium text-secondary-foreground hover:text-brand">
                      <span className="rounded-sm focus-within:ring-2 focus-within:ring-violet-500 focus-within:ring-offset-2">
                        <a
                          href={link.href}
                          className="focus:outline-hidden"
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
                    <p className="text-base text-gray-500">
                      {link.description}
                    </p>
                  </div>
                  <div className="shrink-0 self-center">
                    <ChevronRightIcon
                      className="h-5 w-5 text-gray-400"
                      aria-hidden="true"
                    />
                  </div>
                </li>
              ))}
            </ul>
            <div className="mt-8">
              <a
                href="/"
                className="text-base font-medium text-violet-600 hover:text-violet-500"
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
