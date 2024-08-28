import { ArrowUpRightIcon } from '@heroicons/react/20/solid';
import {
  AcademicCapIcon,
  BookOpenIcon,
  ChatBubbleLeftIcon,
  CloudIcon,
  CodeBracketIcon,
  CommandLineIcon,
  PuzzlePieceIcon,
  StarIcon,
  UsersIcon
} from '@heroicons/react/24/outline';
import { useDispatch } from 'react-redux';
import Button from '~/components/forms/buttons/Button';
import { Icon } from '~/types/Icon';
import { cls } from '~/utils/helpers';
import { useNavigate } from 'react-router-dom';
import { onboardingCompleted } from './events/eventSlice';

const gettingStartedTiles = [
  {
    icon: AcademicCapIcon,
    name: 'Get Started',
    description: 'Learn how to create your first feature flag',
    href: 'https://docs.flipt.io/introduction'
  },
  {
    icon: CloudIcon,
    name: 'Introducing Flipt Cloud',
    className: 'sm:col-span-2',
    description:
      'Learn about our SaaS offering that integrates with your Git repositories and can be used alongside your existing Flipt instances',
    href: 'https://docs.flipt.io/cloud/overview'
  },
  {
    icon: CommandLineIcon,
    name: 'Try the CLI',
    description: 'Use the Flipt CLI to manage your feature flags and more',
    href: 'https://docs.flipt.io/cli/overview'
  },
  {
    icon: BookOpenIcon,
    name: 'Checkout a Guide',
    description:
      'Use Flipt to its full potential. Read our guides including using Flipt with GitOps',
    href: 'https://docs.flipt.io/guides'
  },
  {
    icon: PuzzlePieceIcon,
    name: 'Integrate Your Application',
    description: 'Use our SDKs to integrate your applications in your language',
    href: 'https://docs.flipt.io/integration/overview'
  }
];

const moreTiles = [
  {
    icon: ChatBubbleLeftIcon,
    name: 'Chat With Us',
    description:
      'Join our Discord community to engage with the team and other Flipt users',
    cta: 'Join Discord',
    href: 'https://flipt.io/discord'
  },
  {
    icon: CodeBracketIcon,
    name: 'View API Reference',
    description: 'Learn how to use the Flipt REST API',
    href: 'https://www.flipt.io/docs/reference/overview'
  },
  {
    icon: StarIcon,
    name: 'Support Us',
    description: 'Show your support by starring us on GitHub',
    cta: 'Star Flipt on GitHub',
    href: 'https://github.com/flipt-io/flipt'
  },
  {
    icon: UsersIcon,
    name: 'Join the Community',
    description:
      'Engage with our community on Discourse for support, discussions, and knowledge sharing',
    cta: 'Join Discourse',
    href: 'https://community.flipt.io'
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
}

function OnboardingTile(props: OnboardingTileProps) {
  const {
    className,
    icon: Icon,
    ctaIcon: CTAIcon = ArrowUpRightIcon,
    name,
    description,
    cta,
    href
  } = props;

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
        <h3 className="text-gray-700 text-lg font-semibold">{name}</h3>
        <p className="text-gray-500 max-w-lg">{description}</p>
      </div>
      <div
        className={cls(
          'absolute bottom-0 flex w-full translate-y-10 transform-gpu flex-row items-center p-4 opacity-0 transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100'
        )}
      >
        <a
          href={href || '/#/onboarding'}
          target="_blank"
          rel="noreferrer"
          className="text-gray-500 flex flex-row space-x-1 px-2 py-1 hover:text-gray-700 sm:px-3 sm:py-2"
        >
          <span className="flex">{cta || 'Learn More'}</span>
          <CTAIcon className="my-auto flex h-4 w-4 align-middle" />
        </a>
      </div>
      <div className="pointer-events-none absolute inset-0 transform-gpu transition-all duration-300 group-hover:bg-black/[.03] group-hover:dark:bg-gray-800/10" />
    </div>
  );
}

interface OnboardingProps {
  firstTime?: boolean;
}

export default function Onboarding({ firstTime = false }: OnboardingProps) {
  const navigate = useNavigate();
  const dispatch = useDispatch();

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
        {firstTime && (
          <div className="mt-4">
            <Button
              variant="soft"
              title="Complete Onboarding"
              onClick={() => {
                dispatch(onboardingCompleted());
                navigate('/flags');
              }}
            >
              Continue to Dashboard
            </Button>
          </div>
        )}
      </div>
      <div className="my-8">
        <div className="grid w-full auto-rows-[12rem] grid-cols-1 gap-4 sm:grid-cols-3">
          {gettingStartedTiles.map((tile, i) => (
            <OnboardingTile key={i} {...tile} />
          ))}
        </div>
      </div>
      <div className="mt-12 flex flex-row justify-between pb-5 sm:mt-16 sm:items-center">
        <div className="flex flex-col">
          <h2 className="text-gray-900 text-xl font-bold leading-7 sm:truncate sm:text-2xl">
            More Resources
          </h2>
          <p className="text-gray-500 mt-2 text-sm">
            Once you&apos;re up and running, here are a few more resources
          </p>
        </div>
      </div>
      <div className="my-8">
        <div className="grid w-full auto-rows-[12rem] grid-cols-1 gap-4 sm:grid-cols-3">
          {moreTiles.map((tile, i) => (
            <OnboardingTile key={i} {...tile} />
          ))}
        </div>
      </div>
    </>
  );
}
