import {
  ArrowUpRightIcon,
  BookOpenIcon,
  ChatBubbleBottomCenterIcon,
  EnvelopeIcon,
  ExclamationCircleIcon
} from '@heroicons/react/24/outline';
import React from 'react';
import { Link } from 'react-router-dom';
import { Icon } from '~/types/Icon';

const supportItems: SupportItemProps[] = [
  {
    title: 'Onboarding',
    description: 'Get started with Flipt',
    children: (
      <Link
        to="/onboarding"
        className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-violet-300 sm:px-3 sm:py-2"
      >
        <span className="text-gray-700 text-sm">Let&apos;s Go</span>
      </Link>
    ),
    icon: BookOpenIcon
  },
  {
    title: 'File an Issue',
    description: 'Get support from the community',
    children: (
      <a
        className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-violet-300 sm:px-3 sm:py-2"
        href="https://github.com/flipt-io/flipt/issues/new/choose"
      >
        <span className="text-gray-700 text-sm">Create GitHub Issue</span>
      </a>
    ),
    icon: ExclamationCircleIcon
  },
  {
    title: 'Chat in Discord',
    description: 'Ask a question in our Discord community',
    children: (
      <a
        className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-violet-300 sm:px-3 sm:py-2"
        href="https://www.flipt.io/discord"
      >
        <span className="text-gray-700 text-sm">Join Discord Server</span>
      </a>
    ),
    icon: ChatBubbleBottomCenterIcon
  },
  {
    title: 'Email',
    description: 'Send an email to our shared inbox',
    children: (
      <a
        className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-violet-300 sm:px-3 sm:py-2"
        href="mailto:dev@flipt.io?subject=Support Inquiry"
      >
        <span className="text-gray-700 text-sm">Send Email</span>
      </a>
    ),
    icon: EnvelopeIcon
  }
];

interface SupportItemProps {
  title: string;
  icon: Icon;
  description: string;
  children?: React.ReactNode;
}

function SupportItem(props: SupportItemProps) {
  const { title, description, children, icon: Icon } = props;
  return (
    <div className="border-gray-200 flex h-full w-full flex-col items-stretch space-y-4 rounded-md border p-6">
      <div className="sm:shrink-0">
        <div className="flex items-center space-x-2">
          <Icon className="text-gray-400 h-6 w-6" />
          <h3 className="text-gray-900 text-base font-semibold">{title}</h3>
        </div>
        <p className="text-gray-500 pt-1 text-sm leading-5">{description}</p>
      </div>
      <div className="mt-4 flex grow items-end sm:mt-0">{children}</div>
    </div>
  );
}

export default function Support() {
  return (
    <>
      <div className="flex-row justify-between pb-5 sm:flex sm:items-center">
        <div className="flex flex-col">
          <h1 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl">
            Support
          </h1>
          <p className="text-gray-500 mt-2 text-sm">
            How to get help with Flipt
          </p>
        </div>
        <div className="mt-4">
          <a
            className="text-white bg-violet-500 mb-1 inline-flex items-center justify-center rounded-md border border-transparent px-4 py-2 text-sm font-medium shadow-sm hover:bg-violet-600 hover:cursor-pointer focus:outline-none focus:ring-1 focus:ring-violet-500 focus:ring-offset-1"
            target="_blank"
            rel="noreferrer"
            href="https://www.flipt.io/docs?utm_source=app"
          >
            <span>Documentation</span>
            <ArrowUpRightIcon
              className="text-white -mr-1.5 ml-1 h-3 w-3"
              aria-hidden="true"
            />
          </a>
        </div>
      </div>
      <div className="my-8">
        <div className="container m-auto grid grid-cols-2 gap-4 md:grid-cols-3">
          {supportItems.map((item, index) => (
            <SupportItem key={index} {...item} />
          ))}
        </div>
      </div>
    </>
  );
}
