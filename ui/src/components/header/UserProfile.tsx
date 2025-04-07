import { UserCircleIcon } from '@heroicons/react/24/outline';
import { Button } from '~/components/Button';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { User } from '~/types/auth/User';
import { useNavigate } from 'react-router';
import { expireAuthSelf } from '~/data/api';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator
} from '~/components/ui/dropdown-menu';

type UserProfileProps = {
  user: User;
};

export default function UserProfile(props: UserProfileProps) {
  const { user } = props;
  const { setError } = useError();
  const { clearSession } = useSession();

  const navigate = useNavigate();
  const logout = async () => {
    try {
      await expireAuthSelf();
      clearSession();
      if (user?.issuer) {
        window.location.href = `//${user.issuer}`;
      } else {
        navigate('/login');
      }
    } catch (err) {
      setError(err);
    }
  };

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="link"
          aria-label="Open user menu"
          className="hover:ring-primary/80 focus:ring-primary/80 h-6 w-6 rounded-full ring-2 ring-white ring-offset-0"
        >
          {user.imgURL && (
            <img
              className="h-6 w-6 rounded-full"
              src={user.imgURL}
              alt={user.name}
              title={user.name}
              referrerPolicy="no-referrer"
            />
          )}
          {!user.imgURL && (
            <UserCircleIcon
              aria-hidden="true"
              className="invert dark:invert-0"
              style={{ width: '1.5rem', height: '1.5rem' }}
            />
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {(user.name || user.login) && (
          <>
            <DropdownMenuItem
              disabled
              key="userinfo"
              className="flex flex-col items-start gap-0"
            >
              {user.name}
              {user.login && <span className="text-xs">{user.login}</span>}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem key="logout" onSelect={logout}>
              Logout
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
