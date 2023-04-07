import { PlusIcon } from '@heroicons/react/24/outline';
import { useCallback, useEffect, useRef, useState } from 'react';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/Button';
import Modal from '~/components/Modal';
import ShowTokenPanel from '~/components/settings/tokens/ShowTokenPanel';
import TokenForm from '~/components/settings/tokens/TokenForm';
import TokenTable from '~/components/settings/tokens/TokenTable';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import { deleteToken, listAuthMethods, listTokens } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { IAuthMethod, IAuthMethodList } from '~/types/Auth';
import {
  IAuthToken,
  IAuthTokenInternal,
  IAuthTokenSecret
} from '~/types/auth/Token';

export default function Tokens() {
  // const checkbox = useRef();
  const [tokenAuthEnabled, setTokenAuthEnabled] = useState<boolean>(false);

  const [tokens, setTokens] = useState<IAuthToken[]>([]);

  const [tokensVersion, setTokensVersion] = useState(0);

  const { setError, clearError } = useError();

  const [createdToken, setCreatedToken] = useState<IAuthTokenSecret | null>(
    null
  );
  const [showCreatedTokenModal, setShowCreatedTokenModal] = useState(false);

  const [showTokenForm, setShowTokenForm] = useState<boolean>(false);

  const [showDeleteTokenModal, setShowDeleteTokenModal] =
    useState<boolean>(false);
  const [deletingToken, setDeletingToken] = useState<IAuthToken | null>(null);

  const tokenFormRef = useRef(null);

  const checkTokenAuthEnabled = useCallback(() => {
    listAuthMethods()
      .then((resp: IAuthMethodList) => {
        const authToken = resp.methods.find(
          (m: IAuthMethod) => m.method === 'METHOD_TOKEN' && m.enabled
        );

        setTokenAuthEnabled(!!authToken);
        clearError();
      })
      .catch((err) => {
        setError(err);
      });
  }, [clearError, setError]);

  const fetchTokens = useCallback(() => {
    listTokens()
      .then((data) => {
        const tokens = data.authentications.map((token: IAuthTokenInternal) => {
          return {
            ...token,
            name: token.metadata['io.flipt.auth.token.name'],
            description: token.metadata['io.flipt.auth.token.description']
          };
        });
        setTokens(tokens);
        clearError();
      })
      .catch((err) => {
        setError(err);
      });
  }, [clearError, setError]);

  const incrementTokensVersion = () => {
    setTokensVersion(tokensVersion + 1);
  };

  useEffect(() => {
    fetchTokens();
  }, [tokensVersion, fetchTokens]);

  useEffect(() => {
    checkTokenAuthEnabled();
  }, [checkTokenAuthEnabled]);

  // const [checked, setChecked] = useState(false);
  // const [indeterminate, setIndeterminate] = useState(false);
  // const [selectedTokens, setSelectedTokens] = useState<IAuthenticationToken[]>(
  //   []
  // );

  // useLayoutEffect(() => {
  //   const isIndeterminate =
  //     selectedTokens.length > 0 && selectedTokens.length < tokens.length;
  //   setChecked(selectedTokens.length === tokens.length);
  //   setIndeterminate(isIndeterminate);
  //   if (checkbox && checkbox.current) {
  //     checkbox.current.indeterminate = isIndeterminate;
  //   }
  // }, [selectedTokens]);

  // const toggleAll = () => {
  //   setSelectedTokens(checked || indeterminate ? [] : tokens);
  //   setChecked(!checked && !indeterminate);
  //   setIndeterminate(false);
  // };

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
            incrementTokensVersion();
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
            <>
              Are you sure you want to delete the token{' '}
              <span className="font-medium text-violet-500">
                {deletingToken?.name}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Token"
          setOpen={setShowDeleteTokenModal}
          handleDelete={() => deleteToken(deletingToken?.id ?? '')} // TODO: Determine impact of blank ID param
          onSuccess={() => {
            incrementTokensVersion();
          }}
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
            <h1 className="text-xl font-semibold text-gray-700">
              Static Tokens
            </h1>
            <p className="mt-2 text-sm text-gray-500">
              Static tokens are used to authenticate with the API
            </p>
          </div>
          {tokenAuthEnabled && tokens.length > 0 && (
            <div className="mt-4">
              <Button primary onClick={() => setShowTokenForm(true)}>
                <PlusIcon
                  className="-ml-1.5 mr-1 h-5 w-5 text-white"
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
                setDeletingToken={setDeletingToken}
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
              <p className="text-sm text-gray-600">
                Token Authentication Disabled
              </p>
              <p className="mt-4 text-sm text-gray-500">
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
