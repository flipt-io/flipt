import { ArrowRightIcon, ArrowUpRightIcon } from '@heroicons/react/20/solid';
import {
  AcademicCapIcon,
  BookOpenIcon,
  CloudArrowUpIcon,
  CommandLineIcon,
  LockClosedIcon,
  PuzzlePieceIcon,
  StarIcon
} from '@heroicons/react/24/outline';
import { Icon } from '~/types/Icon';
import { cls } from '~/utils/helpers';

const onboardingTiles = [
  {
    icon: AcademicCapIcon,
    name: 'Get Started',
    description: 'Learn how to create your first feature flag',
    className: 'sm:col-span-2',
    href: 'https://www.flipt.io/docs/introduction'
  },
  {
    icon: CommandLineIcon,
    name: 'Try the CLI',
    description: 'Use the Flipt CLI to manage your feature flags and more',
    href: 'https://www.flipt.io/docs/cli/overview'
  },
  {
    icon: LockClosedIcon,
    name: 'Setup Authentication',
    description: 'Learn how to manage users and service accounts',
    href: 'https://www.flipt.io/docs/authentication/overview'
  },
  {
    icon: PuzzlePieceIcon,
    name: 'Integrate Your Application',
    description: 'Use our SDKs to integrate your applications in your language',
    className: 'sm:col-span-2',
    href: 'https://www.flipt.io/docs/integration/overview'
  },
  {
    icon: CloudArrowUpIcon,
    name: 'Deploy Flipt',
    description: 'Learn how to deploy Flipt in your environment',
    href: 'https://www.flipt.io/docs/operations/deployment'
  },
  {
    icon: BookOpenIcon,
    name: 'Checkout a Guide',
    description:
      'Use Flipt to its full potential. Read our guides including using Flipt with GitOps',
    href: 'https://www.flipt.io/docs/guides'
  },
  {
    icon: StarIcon,
    name: 'Support Us',
    description: 'Show your support by starring us on GitHub',
    cta: 'Star Flipt on GitHub',
    ctaIcon: ArrowUpRightIcon,
    href: 'https://github.com/flipt-io/flipt',
    external: true
  }
];

interface OnboardingTileProps {
  className?: string;
  icon?: Icon;
  name?: string;
  description?: string;
  cta?: string;
  ctaIcon?: Icon;
  href?: string;
  external?: boolean;
}

function OnboardingTile(props: OnboardingTileProps) {
  const {
    className,
    icon: Icon,
    ctaIcon: CTAIcon = ArrowRightIcon,
    name,
    description,
    cta,
    href,
    external
  } = props;

  return (
    <div
      className={cls(
        'group  relative flex flex-col justify-between overflow-hidden rounded-xl hover:border-gray-300 hover:shadow-md hover:shadow-violet-300',
        // light styles
        'bg-white [box-shadow:0_0_0_1px_rgba(0,0,0,.03),0_2px_4px_rgba(0,0,0,.05),0_12px_24px_rgba(0,0,0,.05)]',
        // dark styles
        'transform-gpu dark:bg-transparent dark:backdrop-blur-md dark:[border:1px_solid_rgba(255,255,255,.1)] dark:[box-shadow:0_-20px_80px_-20px_#ffffff1f_inset]',
        className
      )}
    >
      <div className="pointer-events-none z-10 flex transform-gpu flex-col gap-1 p-6 transition-all duration-300 group-hover:-translate-y-10">
        {Icon && (
          <Icon className="text-gray-700 h-6 w-6 origin-left transform-gpu transition-all duration-300 ease-in-out group-hover:scale-50" />
        )}
        <h3 className="text-gray-700 text-lg font-semibold">{name}</h3>
        <p className="text-gray-500 max-w-lg">{description}</p>
      </div>
      <div
        className={cls(
          'absolute bottom-0 flex w-full translate-y-10 transform-gpu flex-row items-center p-4 opacity-0 transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100'
        )}
      >
        {external && (
          <a
            href={href || '/#/onboarding'}
            target="_blank"
            rel="noreferrer"
            className="text-gray-500 flex flex-row space-x-1 px-2 py-1 hover:text-gray-700 sm:px-3 sm:py-2"
          >
            <span className="flex">{cta || 'Learn More'}</span>
            <CTAIcon className="my-auto flex h-4 w-4 align-middle" />
          </a>
        )}
        {!external && (
          <a
            href={href || '/#/onboarding'}
            className="text-gray-500 flex flex-row space-x-1 px-2 py-1 hover:text-gray-700 sm:px-3 sm:py-2"
          >
            <span className="flex">{cta || 'Learn More'}</span>
            <CTAIcon className="my-auto flex h-4 w-4 align-middle" />
          </a>
        )}
      </div>
      <div className="pointer-events-none absolute inset-0 transform-gpu transition-all duration-300 group-hover:bg-black/[.03] group-hover:dark:bg-gray-800/10" />
    </div>
  );
}

export default function Onboarding() {
  return (
    <>
      <div className="flex flex-row justify-between pb-5 sm:items-center">
        <div className="flex flex-col">
          <h1 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl">
            Onboarding
          </h1>
          <p className="text-gray-500 mt-2 text-sm">
            Here are some things to help you get started with Flipt
          </p>
        </div>
      </div>
      <div className="my-8">
        <div className="grid w-full auto-rows-[12rem] grid-cols-1 gap-4 sm:grid-cols-3">
          {onboardingTiles.map((tile, i) => (
            <OnboardingTile key={i} {...tile} />
          ))}
        </div>
      </div>
    </>
  );
}
