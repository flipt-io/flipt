import { Link, useNavigate, useRouteError } from 'react-router-dom';
import logoFlag from '~/assets/logo-flag.png';

export default function ErrorLayout() {
  const error = useRouteError() as Error;
  const navigate = useNavigate();

  return (
    <div className="flex min-h-screen flex-col">
      <main className="mx-auto w-full max-w-7xl px-6 lg:px-8">
        <div className="flex-shrink-0 pt-16">
          <Link to="/">
            <img
              src={logoFlag}
              alt="logo"
              width={512}
              height={512}
              className="m-auto h-20 w-auto"
            />
          </Link>
        </div>
        <div className="mx-auto max-w-xl py-16 sm:py-24">
          <div className="text-center">
            <h1 className="text-gray-900 mt-2 text-4xl font-bold tracking-tight sm:text-5xl">
              Error
            </h1>
            {error && error.message && (
              <p className="text-gray-500 mt-2 text-lg">{error.message}</p>
            )}
          </div>

          <div className="mt-20">
            <a
              href="#"
              onClick={(e) => {
                e.preventDefault();
                navigate(-1);
              }}
              className="text-violet-600 text-base font-medium hover:text-violet-500"
            >
              <span aria-hidden="true">&larr; </span>
              Go Back
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
