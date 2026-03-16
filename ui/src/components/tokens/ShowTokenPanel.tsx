import { Check, Clipboard, Lock } from 'lucide-react';
import { useState } from 'react';
import { Button } from '~/components/Button';
import { IAuthTokenSecret } from '~/types/auth/Token';
import { cls, copyTextToClipboard } from '~/utils/helpers';

type ShowTokenPanelProps = {
  setOpen: (open: boolean) => void;
  token: IAuthTokenSecret | null;
};

export default function ShowTokenPanel(props: ShowTokenPanelProps) {
  const { setOpen, token } = props;
  const [copied, setCopied] = useState(false);
  const [copiedText, setCopiedText] = useState(token?.clientToken);

  return (
    <>
      <div>
        <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
          <Lock className="h-6 w-6 text-green-600" aria-hidden="true" />
        </div>
        <div className="mt-3 text-center sm:mt-5">
          <h3 className="text-secondary-foreground text-lg leading-6 font-medium">
            Created Token
          </h3>
          <div className="mt-2">
            <p className="text-muted-foreground text-sm">
              Please copy the token below and store it in a secure location.
            </p>
            <p className="text-secondary-foreground mt-2 text-sm font-semibold">
              You will NOT be able to view it again
            </p>
          </div>
          <div className="m-auto mt-4 flex content-center bg-[#1a1b26]">
            <div className="m-auto flex">
              <pre className="p-4 text-sm text-[#9aa5ce] md:h-full">
                <code className="text rounded-xs md:h-full">{copiedText}</code>
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
                  <Check
                    className={cls(
                      'absolute m-auto h-6 w-6 justify-center align-middle text-green-400 transition-opacity duration-300 ease-in-out hover:text-white',
                      {
                        'visible opacity-100': copied,
                        'invisible opacity-0': !copied
                      }
                    )}
                  />
                  <Clipboard
                    className={cls(
                      'text-muted-foreground m-auto h-6 w-6 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-white',
                      {
                        'invisible opacity-0': copied,
                        'visible opacity-100': !copied
                      }
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
          variant="primary"
          onClick={() => setOpen(false)}
        >
          Close
        </Button>
      </div>
    </>
  );
}
