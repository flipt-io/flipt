import { Dialog } from '@headlessui/react';
import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';
import hljs from 'highlight.js';
import text from 'highlight.js/lib/languages/plaintext';
import 'highlight.js/styles/tokyo-night-dark.css';
import { useEffect, useState } from 'react';
import Button from '~/components/forms/buttons/Button';
import { IAuthTokenSecret } from '~/types/auth/Token';
import { classNames, copyTextToClipboard } from '~/utils/helpers';

hljs.registerLanguage('text', text);

type ShowTokenPanelProps = {
  setOpen: (open: boolean) => void;
  token: IAuthTokenSecret | null;
};

export default function ShowTokenPanel(props: ShowTokenPanelProps) {
  const { setOpen, token } = props;
  const [copied, setCopied] = useState(false);
  const [copiedText, setCopiedText] = useState(token?.clientToken);

  useEffect(() => {
    hljs.initHighlighting();
  }, [token]);

  return (
    <>
      <div>
        <div className="bg-green-100 mx-auto flex h-12 w-12 items-center justify-center rounded-full">
          <CheckIcon className="text-green-600 h-6 w-6" aria-hidden="true" />
        </div>
        <div className="mt-3 text-center sm:mt-5">
          <Dialog.Title
            as="h3"
            className="text-gray-900 text-lg font-medium leading-6"
          >
            Created Token
          </Dialog.Title>
          <div className="mt-2">
            <p className="text-gray-500 text-sm">
              Please copy the token below and store it in a secure location.
            </p>
            <p className="text-gray-700 mt-2 text-sm">
              You will not be able to view it again
            </p>
          </div>
          <div className="m-auto mt-4 flex content-center bg-[#1a1b26]">
            <div className="m-auto flex">
              <pre className="p-2 text-sm md:h-full">
                <code className="text rounded-sm md:h-full">{copiedText}</code>
              </pre>
              {token?.clientToken && (
                <button
                  aria-label="Copy"
                  className="hidden md:block"
                  onClick={() => {
                    copyTextToClipboard(token?.clientToken);
                    setCopied(true);
                    setCopiedText('Copied to clipboard');
                    setTimeout(() => {
                      setCopied(false);
                      setCopiedText(token?.clientToken);
                    }, 2000);
                  }}
                >
                  <CheckIcon
                    className={classNames(
                      'text-green-400 absolute m-auto h-6 w-6 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-white',
                      copied ? 'visible opacity-100' : 'invisible opacity-0'
                    )}
                  />
                  <ClipboardDocumentIcon
                    className={classNames(
                      'text-gray-400 m-auto h-6 w-6 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-white',
                      copied ? 'invisible opacity-0' : 'visible opacity-100'
                    )}
                  />
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
      <div className="mt-5 sm:mt-6">
        <Button
          className="inline-flex justify-center sm:w-full"
          primary
          onClick={() => setOpen(false)}
        >
          Close
        </Button>
      </div>
    </>
  );
}
