import { QuestionMarkCircleIcon } from '@heroicons/react/20/solid';

type MoreInfoProps = {
  className?: string;
  href: string;
  children: React.ReactNode;
};

export default function MoreInfo(props: MoreInfoProps) {
  const { className, href, children } = props;

  return (
    <div className={`${className} flex text-xs tracking-tight`}>
      <a
        href={href}
        target="_blank"
        rel="noreferrer"
        className="group inline-flex items-center text-gray-500 hover:text-gray-600"
      >
        <QuestionMarkCircleIcon
          className="-ml-1 h-4 w-4 text-gray-300 group-hover:text-gray-400"
          aria-hidden="true"
        />
        <span className="ml-1">{children}</span>
      </a>
    </div>
  );
}
