import {
  BookOpenIcon,
  CircleAlertIcon,
  ExternalLinkIcon,
  GraduationCapIcon,
  MailIcon,
  MessageCircle,
  PuzzleIcon,
  StarIcon,
  UsersIcon
} from 'lucide-react';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router';

import { Button } from '~/components/Button';

import { Icon } from '~/types/Icon';

import { cls } from '~/utils/helpers';

import { onboardingCompleted } from './events/eventSlice';

const gettingStartedTiles = [
  {
    icon: GraduationCapIcon,
    name: 'Quick Start',
    description: 'Learn how to create your first feature flag',
    href: 'https://docs.flipt.io/v2/quickstart'
  },
  {
    icon: BookOpenIcon,
    name: 'Checkout a Guide',
    description:
      'Use Flipt to its full potential. Read our guides using Flipt with GitOps',
    href: 'https://docs.flipt.io/v2/guides'
  },
  {
    icon: PuzzleIcon,
    name: 'Integrate Your Application',
    description: 'Use our SDKs to integrate your applications in your language',
    href: 'https://docs.flipt.io/v2/integration/overview'
  }
];

const moreTiles = [
  {
    icon: MessageCircle,
    name: 'Chat With Us',
    description:
      'Join our Discord community to engage with the team and other Flipt users',
    cta: 'Join Discord',
    href: 'https://flipt.io/discord'
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
      'Engage with our community on GitHub for support, discussions, and knowledge sharing',
    cta: 'Join GitHub',
    href: 'https://github.com/flipt-io/flipt/discussions'
  },
  {
    icon: MailIcon,
    name: 'Email',
    description: 'Send an email to our shared inbox',
    cta: 'Send Email',
    href: 'mailto:dev@flipt.io?subject=Support Inquiry'
  },
  {
    icon: CircleAlertIcon,
    name: 'Report an issue',
    description: 'Spotted something? Want something? Let us know!',
    cta: 'Create GitHub Issue',
    href: 'https://github.com/flipt-io/flipt/issues/new/choose'
  }
];

interface SupportTileProps {
  className?: string;
  icon?: Icon;
  name?: string;
  description?: string;
  cta?: string;
  ctaIcon?: Icon;
  href?: string;
}

function SupportTile(props: SupportTileProps) {
  const {
    className,
    icon: Icon,
    ctaIcon: CTAIcon = ExternalLinkIcon,
    name,
    description,
    cta,
    href
  } = props;

  return (
    <div
      className={cls(
        'group relative border flex flex-col justify-between overflow-hidden rounded-xl bg-secondary/20 dark:bg-secondary/50 hover:bg-accent/50',
        className
      )}
    >
      <div className="pointer-events-none z-10 flex transform-gpu flex-col gap-1 p-6 transition-all duration-300 group-hover:-translate-y-10 ">
        {Icon && (
          <Icon className="h-6 w-6 origin-left transform-gpu text-muted-foreground " />
        )}
        <h3 className="text-lg font-semibold text-secondary-foreground">
          {name}
        </h3>
        <p className="max-w-lg text-muted-foreground">{description}</p>
      </div>
      <div
        tabIndex={0}
        className={cls(
          'absolute bottom-0 flex w-full sm:translate-y-10 transform-gpu flex-row items-center p-4 sm:opacity-0 transition-all duration-300 group-hover:opacity-100 group-hover:translate-y-0'
        )}
      >
        <a
          href={href || '/#/onboarding'}
          target="_blank"
          rel="noreferrer"
          className="flex flex-row space-x-1 px-2 py-1 text-foreground hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent sm:px-3 sm:py-2 rounded-md"
        >
          <span className="flex">{cta || 'Learn More'}</span>
          <CTAIcon className="my-auto flex h-4 w-4 align-middle" />
        </a>
      </div>
      <div className="pointer-events-none absolute inset-0 transform-gpu transition-all duration-300 " />
    </div>
  );
}

interface SupportProps {
  firstTime?: boolean;
}

export default function Support({ firstTime = false }: SupportProps) {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const title = firstTime ? 'Onboarding' : 'Support';

  return (
    <>
      <div className="flex flex-row justify-between pb-5 sm:items-center">
        <div className="flex flex-col">
          <h1 className="text-2xl font-bold leading-7 sm:leading-9 text-foreground sm:truncate sm:text-3xl ">
            {title}
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
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
        <div className="grid w-full auto-rows-[12rem] grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {gettingStartedTiles.map((tile, i) => (
            <SupportTile key={i} {...tile} />
          ))}
        </div>
      </div>
      <div className="mt-12 flex flex-row justify-between pb-5 sm:mt-16 sm:items-center">
        <div className="flex flex-col">
          <h2 className="text-xl font-bold leading-7 text-gray-900 dark:text-gray-100 sm:truncate sm:text-2xl">
            More Resources
          </h2>
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
            Once you&apos;re up and running, here are a few more resources
          </p>
        </div>
      </div>
      <div className="my-8">
        <div className="grid w-full auto-rows-[12rem] grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {moreTiles.map((tile, i) => (
            <SupportTile key={i} {...tile} />
          ))}
        </div>
      </div>
    </>
  );
}
