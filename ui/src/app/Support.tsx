import {
  ArrowUpRightIcon,
  ChatBubbleBottomCenterIcon,
  EnvelopeIcon,
  ExclamationCircleIcon
} from '@heroicons/react/24/outline';

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
            href="https://www.flipt.io/docs/"
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
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <div className="flex h-72 flex-col space-y-4 sm:flex-row sm:space-x-4 sm:space-y-0 lg:h-48">
              <div className="border-gray-200 flex h-full w-full flex-col items-stretch rounded-md border p-6 sm:w-1/3">
                <div className="sm:shrink-0">
                  <div className="flex items-center space-x-2">
                    <ExclamationCircleIcon className="text-gray-400 h-6 w-6" />
                    <h3 className="text-gray-900 text-base font-semibold">
                      File an Issue
                    </h3>
                  </div>
                  <p className="text-gray-500 pt-1 text-sm leading-5">
                    Get support from the entire community
                  </p>
                </div>
                <div className="mt-4 flex grow items-end sm:mt-0">
                  <a
                    className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-gray-400 sm:px-3 sm:py-2"
                    href="https://github.com/flipt-io/flipt/issues/new/choose"
                  >
                    <span className="text-gray-700 text-sm">
                      Create GitHub Issue
                    </span>
                  </a>
                </div>
              </div>
              <div className="border-gray-200 flex h-full w-full flex-col items-stretch rounded-md border p-6 sm:w-1/3">
                <div className="sm:shrink-0">
                  <div className="flex items-center space-x-2">
                    <ChatBubbleBottomCenterIcon className="text-gray-400 h-6 w-6" />
                    <h3 className="text-gray-900 text-base font-semibold">
                      Slack Connect
                    </h3>
                  </div>
                  <p className="text-gray-500 pt-1 text-sm leading-5">
                    Invite your team to collaborate in a shared Slack channel
                  </p>
                </div>
                <div className="mt-4 flex grow items-end sm:mt-0">
                  <a
                    className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-gray-400 sm:px-3 sm:py-2"
                    href="mailto:info@flipt.io?subject=Slack Connect Request&body=Hi there! I'd like to request a Slack Connect channel for our team."
                  >
                    <span className="text-gray-700 text-sm">
                      Request Invite
                    </span>
                  </a>
                </div>
              </div>
              <div className="border-gray-200 flex h-full w-full flex-col items-stretch rounded-md border p-6 sm:w-1/3">
                <div className="sm:shrink-0">
                  <div className="flex items-center space-x-2">
                    <EnvelopeIcon className="text-gray-400 h-6 w-6" />
                    <h3 className="text-gray-900 text-base font-semibold">
                      Email
                    </h3>
                  </div>
                  <p className="text-gray-500 pt-1 text-sm leading-5">
                    Send an email to our shared inbox
                  </p>
                </div>
                <div className="mt-4 flex grow items-end sm:mt-0">
                  <a
                    className="border-gray-200 rounded-md border px-2 py-1 hover:border-gray-300 hover:shadow-sm hover:shadow-gray-400 sm:px-3 sm:py-2"
                    href="mailto:info@flipt.io?subject=Support Inquiry"
                  >
                    <span className="text-gray-700 text-sm">Send Email</span>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
