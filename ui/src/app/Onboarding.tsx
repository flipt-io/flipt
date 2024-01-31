import {
  AcademicCapIcon,
  BookOpenIcon,
  CloudArrowUpIcon,
  Cog6ToothIcon,
  LockClosedIcon,
  PuzzlePieceIcon,
  StarIcon
} from '@heroicons/react/24/outline';
import { Icon } from '~/types/Icon';
import { cls } from '~/utils/helpers';

const onboardingTiles = [
  {
    Icon: AcademicCapIcon,
    name: 'Get Started',
    description: 'Learn how to create your first feature flag',
    className: 'sm:col-span-2'
  },
  {
    Icon: Cog6ToothIcon,
    name: 'Configure Flipt',
    description: 'Setup Flipt to fit your needs'
  },
  {
    Icon: LockClosedIcon,
    name: 'Setup Authentication',
    description: 'Learn how to manage users and service accounts'
  },
  {
    Icon: PuzzlePieceIcon,
    name: 'Integrate Your Application',
    description: 'Use our SDKs to integrate your apps',
    className: 'sm:col-span-2'
  },
  {
    Icon: CloudArrowUpIcon,
    name: 'Deploy Flipt',
    description: 'Learn how to deploy Flipt in your environment'
  },
  {
    Icon: BookOpenIcon,
    name: 'Checkout a Guide',
    description:
      'Use Flipt to its full potential. Read our guides including using Flipt with GitOps'
  },
  {
    Icon: StarIcon,
    name: 'Support Us',
    description: 'Show your support by starring us on GitHub'
  }
];

interface OnboardingTileProps {
  className?: string;
  Icon?: Icon;
  name?: string;
  description?: string;
  cta?: string;
  href?: string;
}

function OnboardingTile(props: OnboardingTileProps) {
  const { className, Icon, name, description, cta, href } = props;

  return (
    <div
      className={cls(
        'group relative flex flex-col justify-between overflow-hidden rounded-xl hover:border-gray-300 hover:shadow-md hover:shadow-violet-300',
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
        <h3 className="text-gray-700 text-xl font-semibold dark:text-gray-300">
          {name}
        </h3>
        <p className="text-gray-400 max-w-lg">{description}</p>
      </div>
      <div
        className={cls(
          'absolute bottom-0 flex w-full translate-y-10 transform-gpu flex-row items-center p-4 opacity-0 transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100'
        )}
      >
        <a
          href={href || '#'}
          className="text-gray-500 px-2 py-1 hover:text-gray-700 sm:px-3 sm:py-2"
        >
          {cta || 'Learn More'}
        </a>
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
            Onboarding Guide
          </h1>
          <p className="text-gray-500 mt-2 text-sm">
            Welcome to Flipt! Here are some things to help you get started
          </p>
        </div>
      </div>
      <div className="my-8">
        <div className="grid w-full auto-rows-[14rem] grid-cols-1 gap-4 sm:grid-cols-3">
          {onboardingTiles.map((tile, i) => (
            <OnboardingTile key={i} {...tile} />
          ))}
        </div>
      </div>
    </>
  );
}
