import { CheckIcon, ClipboardCopyIcon } from 'lucide-react';
import React from 'react';

import { cls, copyTextToClipboard } from '~/utils/helpers';

export interface ClipboardCopyProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  text: string;
}

const ClipboardCopy = React.forwardRef<HTMLButtonElement, ClipboardCopyProps>(
  ({ className, ...props }) => {
    const [keyCopied, setKeyCopied] = React.useState(false);
    return (
      <button
        aria-label="Copy to clipboard"
        title="Copy to Clipboard"
        className={cls('hidden md:block', className)}
        onClick={(e) => {
          e.preventDefault();
          copyTextToClipboard(props.text);
          setKeyCopied(true);
          setTimeout(() => {
            setKeyCopied(false);
          }, 2000);
        }}
      >
        <CheckIcon
          className={cls(
            'absolute m-auto h-5 w-5 justify-center align-middle text-green-400 transition-opacity duration-300 ease-in-out',
            {
              'visible opacity-100': keyCopied,
              'invisible opacity-0': !keyCopied
            }
          )}
        />
        <ClipboardCopyIcon
          className={cls(
            'm-auto h-5 w-5 justify-center align-middle text-gray-300 transition-opacity duration-300 ease-in-out hover:text-gray-400',
            {
              'visible opacity-100': !keyCopied,
              'invisible opacity-0': keyCopied
            }
          )}
        />
      </button>
    );
  }
);

ClipboardCopy.displayName = 'ClipboardCopy';
export default ClipboardCopy;
