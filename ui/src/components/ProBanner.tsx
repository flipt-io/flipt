import { SparklesIcon, XIcon } from 'lucide-react';
import { useDispatch } from 'react-redux';

import { proBannerDismissed } from '~/app/events/eventSlice';

export default function ProBanner() {
  const dispatch = useDispatch();

  return (
    <div className="flex items-center gap-x-6 bg-gradient-to-r from-purple-600 to-indigo-600 px-6 py-2.5 sm:px-3.5 sm:before:flex-1">
      <p className="text-sm leading-6 text-white">
        <a href="https://docs.flipt.io/v2/pro">
          <strong className="font-semibold">Introducing Flipt Pro</strong>
          <SparklesIcon className="mx-2 mb-1 inline h-4 w-4" />
          Create pull requests directly from Flipt, GPG signing, secrets
          management, and more.{' '}
          <span className="underline underline-offset-4">
            Start your 14-day free trial!
          </span>
          &nbsp;
          <span aria-hidden="true">&rarr;</span>
        </a>
      </p>
      <div className="flex flex-1 justify-end">
        <button
          type="button"
          className="-m-3 p-3 focus-visible:outline-offset-[-4px]"
          onClick={() => dispatch(proBannerDismissed())}
        >
          <span className="sr-only">Dismiss</span>
          <XIcon className="h-5 w-5 text-white" aria-hidden="true" />
        </button>
      </div>
    </div>
  );
}
