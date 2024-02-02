import { PlusIcon } from '@heroicons/react/24/outline';
import { useMemo, useRef, useState } from 'react';
import { useListAuthProvidersQuery } from '~/app/auth/authApi';
import {
  useDeleteTokensMutation,
  useListTokensQuery
} from '~/app/tokens/tokensApi';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import Slideover from '~/components/Slideover';
import ShowTokenPanel from '~/components/tokens/ShowTokenPanel';
import TokenForm from '~/components/tokens/TokenForm';
import TokenTable from '~/components/tokens/TokenTable';
import Well from '~/components/Well';
import { useError } from '~/data/hooks/error';
import { IAuthMethod } from '~/types/Auth';
import {
  IAuthToken,
  IAuthTokenInternal,
  IAuthTokenSecret
} from '~/types/auth/Token';

export default function Tokens() {
  const { setError } = useError();

  const [createdToken, setCreatedToken] = useState<IAuthTokenSecret | null>(
    null
  );
  const [showCreatedTokenModal, setShowCreatedTokenModal] = useState(false);

  const [showTokenForm, setShowTokenForm] = useState<boolean>(false);

  const [showDeleteTokenModal, setShowDeleteTokenModal] =
    useState<boolean>(false);
  const [deletingTokens, setDeletingTokens] = useState<IAuthToken[] | null>(
    null
  );

  const [deleteTokens] = useDeleteTokensMutation();

  const tokenFormRef = useRef(null);

  const { data: listAuthProviders } = useListAuthProvidersQuery();

  const tokenAuthEnabled = useMemo(() => {
    return listAuthProviders?.methods.find(
      (m: IAuthMethod) => m.method === 'METHOD_TOKEN' && m.enabled
    );
  }, [listAuthProviders]);

  const tokensQuery = useListTokensQuery();

  const tokens = useMemo(
    () =>
      (tokensQuery.data?.authentications || []).map(
        (token: IAuthTokenInternal) => {
          return {
            ...token,
            name: token.metadata['io.flipt.auth.token.name'],
            description:
              token.metadata['io.flipt.auth.token.description'] ?? '',
            namespaceKey: token.metadata['io.flipt.auth.token.namespace'] ?? ''
          };
        }
      ),
    [tokensQuery]
  );

  if (tokensQuery.isError) {
    setError(tokensQuery.error);
  }

  if (tokensQuery.isLoading) {
    return <Loading />;
  }

  return (
    <>
      {/* token create form */}
      <Slideover
        open={showTokenForm}
        setOpen={setShowTokenForm}
        ref={tokenFormRef}
      >
        <TokenForm
          ref={tokenFormRef}
          setOpen={setShowTokenForm}
          onSuccess={(token: IAuthTokenSecret) => {
            setShowTokenForm(false);
            setCreatedToken(token);
            setShowCreatedTokenModal(true);
          }}
        />
      </Slideover>

      {/* token delete modal */}
      <Modal open={showDeleteTokenModal} setOpen={setShowDeleteTokenModal}>
        <DeletePanel
          panelMessage={
            deletingTokens && deletingTokens.length === 1 ? (
              <>
                Are you sure you want to delete the token{' '}
                <span className="text-violet-500 font-medium">
                  {deletingTokens[0].name}
                </span>
                ? This action cannot be undone.
              </>
            ) : (
              <>
                Are you sure you want to delete the selected tokens? This action
                cannot be undone.
              </>
            )
          }
          panelType="Tokens"
          setOpen={setShowDeleteTokenModal}
          handleDelete={() =>
            deleteTokens(deletingTokens?.map((t) => t.id) || [])
          }
        />
      </Modal>

      {/* token created modal */}
      <Modal open={showCreatedTokenModal} setOpen={setShowCreatedTokenModal}>
        <ShowTokenPanel
          token={createdToken}
          setOpen={setShowCreatedTokenModal}
        />
      </Modal>

      <div className="my-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h3 className="text-gray-700 text-xl font-semibold">
              Static Tokens
            </h3>
            <p className="text-gray-500 mt-2 text-sm">
              Static tokens are used to authenticate with the API
            </p>
          </div>
          {tokenAuthEnabled && tokens?.length > 0 && (
            <div className="mt-4">
              <Button variant="primary" onClick={() => setShowTokenForm(true)}>
                <PlusIcon
                  className="text-white -ml-1.5 mr-1 h-5 w-5"
                  aria-hidden="true"
                />
                <span>New Token</span>
              </Button>
            </div>
          )}
        </div>
        {tokenAuthEnabled ? (
          <div className="mt-8 flex flex-col">
            {tokens && tokens.length > 0 ? (
              <TokenTable
                tokens={tokens}
                setDeletingTokens={setDeletingTokens}
                setShowDeleteTokenModal={setShowDeleteTokenModal}
              />
            ) : (
              <EmptyState
                text="New Token"
                onClick={() => {
                  setShowTokenForm(true);
                }}
              />
            )}
          </div>
        ) : (
          <div className="mt-8 flex flex-col text-center">
            <Well>
              <p className="text-gray-600 text-sm">
                Token Authentication Disabled
              </p>
              <p className="text-gray-500 mt-4 text-sm">
                See the configuration{' '}
                <a
                  className="text-violet-500"
                  href="https://www.flipt.io/docs/configuration/authentication"
                >
                  documentation
                </a>{' '}
                for more information.
              </p>
            </Well>
          </div>
        )}
      </div>
    </>
  );
}
