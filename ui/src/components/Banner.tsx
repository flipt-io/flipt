import { XMarkIcon } from '@heroicons/react/20/solid';
import { useDispatch } from 'react-redux';
import { bannerDismissed } from '~/app/events/eventSlice';

type BannerProps = {
  title: string;
  description: string;
  href: string;
};

export default function Banner(props: BannerProps) {
  const { title, description, href } = props;

  const dispatch = useDispatch();

  return (
    <div className="bg-violet-700 flex items-center gap-x-6 px-6 py-2.5 sm:px-3.5 sm:before:flex-1">
      <p className="text-white text-sm leading-6">
        <a href={href}>
          <strong className="font-semibold">{title}</strong>
          <svg
            viewBox="0 0 2 2"
            className="mx-2 inline h-0.5 w-0.5 fill-current"
            aria-hidden="true"
          >
            <circle cx={1} cy={1} r={1} />
          </svg>
          {description}&nbsp;
          <span aria-hidden="true">&rarr;</span>
        </a>
      </p>
      <div className="flex flex-1 justify-end">
        <button
          type="button"
          className="-m-3 p-3 focus-visible:outline-offset-[-4px]"
          onClick={() => dispatch(bannerDismissed())}
        >
          <span className="sr-only">Dismiss</span>
          <XMarkIcon className="text-white h-5 w-5" aria-hidden="true" />
        </button>
      </div>
    </div>
  );
}
