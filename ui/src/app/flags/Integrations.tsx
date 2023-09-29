import { ArrowUpRightIcon } from '@heroicons/react/24/outline';
import { useOutletContext } from 'react-router-dom';
import Well from '~/components/Well';
import { FlagProps } from './FlagProps';

const actions = [
  {
    title: 'REST SDKs',
    href: 'integration/rest',
    description: 'Use our official REST SDKs or generate your own.'
  },
  {
    title: 'GRPC SDKs',
    href: 'integration/grpc',
    description: 'Use our official GRPC SDKs or generate your own.'
  },
  {
    title: 'OpenFeature',
    href: 'integration/openfeature',
    description:
      'OpenFeature is an open specification that provides a vendor-agnostic, community-driven API for feature flagging that works with your favorite feature flag management tool.'
  }
];

export default function Integrations() {
  const { flag } = useOutletContext<FlagProps>();

  return (
    <div className="flex-row justify-between pb-5 sm:flex sm:items-center">
      <div className="flex w-full flex-col">
        <p className="text-gray-500 mt-5 text-sm">
          How to integrate <span className="text-gray-700">{flag.key}</span>{' '}
          into your applications.
        </p>
        <Well className="mt-10 w-full">
          <div className="divide-y divide-gray-200 overflow-hidden sm:grid sm:grid-cols-2 sm:gap-px sm:divide-y-0">
            {actions.map((action) => (
              <div
                key={action.title}
                className="group bg-white relative p-6 focus-within:ring-1 focus-within:ring-inset focus-within:ring-violet-400"
              >
                <div className="my-4">
                  <h3 className="text-gray-700 text-base font-semibold leading-6">
                    <a
                      target="_blank"
                      rel="noreferrer"
                      href={`https://www.flipt.io/docs/${action.href}?utm_source=app`}
                      className="focus:outline-none"
                    >
                      {/* Extend touch target to entire panel */}
                      <span className="absolute inset-0" aria-hidden="true" />
                      {action.title}
                    </a>
                  </h3>
                  <p className="text-gray-500 mt-2 text-sm">
                    {action.description}
                  </p>
                </div>
                <span
                  className="text-gray-300 pointer-events-none absolute right-6 top-6 group-hover:text-gray-400"
                  aria-hidden="true"
                >
                  <ArrowUpRightIcon className="h-4 w-4" aria-hidden="true" />
                </span>
              </div>
            ))}
          </div>
        </Well>
      </div>
    </div>
  );
}
