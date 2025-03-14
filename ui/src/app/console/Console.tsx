import { json } from '@codemirror/lang-json';
import { CodeIcon, SquareTerminalIcon, RefreshCwIcon } from 'lucide-react';
import { tokyoNight } from '@uiw/codemirror-theme-tokyo-night';
import CodeMirror from '@uiw/react-codemirror';
import { Form, Formik, useFormikContext } from 'formik';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import { v4 as uuidv4 } from 'uuid';
import * as Yup from 'yup';
import { useListAuthProvidersQuery } from '~/app/auth/authApi';
import { useListFlagsQuery } from '~/app/flags/flagsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { selectCurrentRef } from '~/app/refs/refsSlice';
import { JsonEditor } from '~/components/json/JsonEditor';
import EmptyState from '~/components/EmptyState';
import { Button } from '~/components/Button';
import Combobox from '~/components/Combobox';
import Dropdown from '~/components/Dropdown';
import Input from '~/components/forms/Input';
import { evaluateURL, evaluateV2 } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import {
  contextValidation,
  keyValidation,
  requiredValidation
} from '~/data/validations';
import { IAuthMethod } from '~/types/Auth';
import { Command } from '~/types/Cli';
import { FilterableFlag, FlagType, flagTypeToLabel, IFlag } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';
import {
  copyTextToClipboard,
  generateCliCommand,
  generateCurlCommand,
  getErrorMessage
} from '~/utils/helpers';
import { PageHeader } from '~/components/ui/page';

function ResetOnNamespaceChange({ namespace }: { namespace: INamespace }) {
  const { resetForm } = useFormikContext();

  useEffect(() => {
    resetForm();
  }, [namespace, resetForm]);

  return null;
}

const consoleValidationSchema = Yup.object({
  flagKey: keyValidation,
  entityId: requiredValidation,
  context: contextValidation
});

interface ConsoleFormValues {
  flagKey: string;
  entityId: string;
  context: string;
}

export default function Console() {
  const [selectedFlag, setSelectedFlag] = useState<FilterableFlag | null>(null);
  const [response, setResponse] = useState<string | null>(null);
  const [hasEvaluationError, setHasEvaluationError] = useState<boolean>(false);

  const { setError, clearError } = useError();
  const navigate = useNavigate();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);
  const ref = useSelector(selectCurrentRef);
  const { data, error } = useListFlagsQuery(namespace.key);

  useEffect(() => {
    if (error) {
      setError(error);
      return;
    }
    clearError();
  }, [clearError, error, setError]);

  const flags = useMemo(() => {
    const initialFlags = data?.flags || [];
    return initialFlags.map((flag) => {
      const status =
        flag.enabled || flag.type === FlagType.BOOLEAN ? 'active' : 'inactive';

      return {
        ...flag,
        status: status as 'active' | 'inactive',
        filterValue: flag.key,
        displayValue: `${flag.name} | ${flagTypeToLabel(flag.type)}`
      };
    });
  }, [data]);

  const { data: listAuthProviders } = useListAuthProvidersQuery();

  const isAuthRequired = useMemo(() => {
    return (
      (listAuthProviders?.methods || []).filter((m: IAuthMethod) => m.enabled)
        .length > 0
    );
  }, [listAuthProviders]);

  const handleSubmit = (flag: IFlag | null, values: ConsoleFormValues) => {
    const { entityId, context } = values;

    if (!flag) {
      return;
    }

    let parsed = null;
    try {
      // need to unescape the context string
      parsed = JSON.parse(context);
    } catch (err) {
      setHasEvaluationError(true);
      setResponse('error: ' + getErrorMessage(err));
      return;
    }

    const rest = {
      entityId,
      context: parsed
    };

    evaluateV2(ref, namespace.key, flag.key, flag.type, rest)
      .then((resp) => {
        setHasEvaluationError(false);
        setResponse(JSON.stringify(resp, null, 2));
      })
      .catch((err) => {
        setHasEvaluationError(true);
        setResponse('error: ' + err.message);
      });
  };

  const handleCopyAsCli = (values: ConsoleFormValues) => {
    let parsed = null;
    try {
      // need to unescape the context string
      parsed = JSON.parse(values.context);
    } catch (err) {
      setHasEvaluationError(true);
      setError('Context provided is invalid.');
      return;
    }

    const contextOptions = Object.entries(parsed).map(([key, value]) => ({
      key: '--context',
      value: `${key}=${value}`
    }));

    const command = generateCliCommand({
      commandName: Command.Evaluate,
      arguments: [values.flagKey],
      options: [
        { key: '--entity-id', value: values.entityId },
        { key: '--namespace', value: namespace.key },
        ...contextOptions
      ]
    });
    copyTextToClipboard(command);
    setSuccess('Command copied to clipboard');
  };

  const handleCopyAsCurl = (values: ConsoleFormValues) => {
    let parsed = null;
    try {
      // need to unescape the context string
      parsed = JSON.parse(values.context);
    } catch (err) {
      setHasEvaluationError(true);
      setError('Context provided is invalid.');
      return;
    }
    const uri =
      window.location.origin +
      '/' +
      evaluateURL +
      (selectedFlag?.type === FlagType.BOOLEAN ? '/boolean' : '/variant');

    let headers: Record<string, string> = {};

    if (isAuthRequired) {
      // user can generate an auth token and use it
      headers.Authorization = 'Bearer <api-token>';
    }

    const command = generateCurlCommand({
      method: 'POST',
      body: {
        ...values,
        context: parsed,
        namespaceKey: namespace.key
      },
      headers,
      uri
    });

    copyTextToClipboard(command);

    setSuccess('Command copied to clipboard');
  };

  const initialvalues: ConsoleFormValues = {
    flagKey: selectedFlag?.key || '',
    entityId: uuidv4(),
    context: '{}'
  };

  return (
    <>
      <PageHeader title="Console" />
      <p className="mt-2 text-sm text-gray-500">
        See the results of your flag evaluations and debug any issues
      </p>
      <div className="flex flex-col md:flex-row">
        {flags.length > 0 && (
          <>
            <div className="mt-8 w-full overflow-hidden md:w-1/2">
              <Formik
                initialValues={initialvalues}
                validationSchema={consoleValidationSchema}
                onSubmit={(values) => {
                  handleSubmit(selectedFlag, values);
                }}
              >
                {(formik) => (
                  <Form className="px-1 sm:overflow-hidden sm:rounded-md">
                    <div className="space-y-6">
                      <div className="grid grid-cols-3 gap-6">
                        <div className="col-span-3">
                          <label
                            htmlFor="flagKey"
                            className="block text-sm font-medium text-gray-700"
                          >
                            Flag Key
                          </label>
                          <Combobox<FilterableFlag>
                            id="flagKey"
                            name="flagKey"
                            className="mt-1"
                            placeholder="Select or search for a flag"
                            values={flags}
                            selected={selectedFlag}
                            setSelected={setSelectedFlag}
                          />
                        </div>
                        <div className="col-span-3">
                          <label
                            htmlFor="entityId"
                            className="block text-sm font-medium text-gray-700"
                          >
                            Entity ID
                          </label>
                          <div className="flex items-center justify-between">
                            <Input
                              className="mt-1 md:mr-2"
                              name="entityId"
                              id="entityId"
                              type="text"
                            />
                            <Button
                              aria-label="New Entity ID"
                              title="New Entity ID"
                              variant="ghost"
                              onClick={(e) => {
                                e.preventDefault();
                                formik.setFieldValue('entityId', uuidv4());
                              }}
                            >
                              <RefreshCwIcon className="h-4 w-4 text-gray-400" />
                            </Button>
                          </div>
                        </div>
                        <div className="col-span-3">
                          <label
                            htmlFor="context"
                            className="block text-sm font-medium text-gray-700"
                          >
                            Request Context
                          </label>
                          <div className="mt-1 text-sm">
                            <JsonEditor
                              id="context"
                              value={formik.values.context}
                              setValue={(v) => {
                                formik.setFieldValue('context', v);
                              }}
                            />
                          </div>
                        </div>
                      </div>
                      <div className="flex justify-end gap-2">
                        <Dropdown
                          disabled={!(formik.dirty && formik.isValid)}
                          label="Copy"
                          side="top"
                          actions={[
                            {
                              id: 'curl',
                              disabled: !(formik.dirty && formik.isValid),
                              label: 'Curl Request',
                              onClick: () => handleCopyAsCurl(formik.values),
                              icon: CodeIcon
                            },
                            {
                              id: 'cli',
                              disabled: !(formik.dirty && formik.isValid),
                              label: 'Flipt CLI',
                              onClick: () => handleCopyAsCli(formik.values),
                              icon: SquareTerminalIcon
                            }
                          ]}
                        />
                        <Button
                          variant="primary"
                          type="submit"
                          disabled={!(formik.dirty && formik.isValid)}
                        >
                          Evaluate
                        </Button>
                      </div>
                    </div>
                    <ResetOnNamespaceChange namespace={namespace} />
                  </Form>
                )}
              </Formik>
            </div>
            <div className="mt-10 w-full overflow-hidden md:w-1/2 md:pl-4">
              {response && (
                <pre className="bg-[#1a1b26] p-2 text-sm md:h-full">
                  {hasEvaluationError ? (
                    <p className="text-red-400">{response}</p>
                  ) : (
                    <CodeMirror
                      value={response}
                      height="100%"
                      extensions={[json()]}
                      basicSetup={{
                        lineNumbers: false,
                        foldGutter: false,
                        highlightActiveLine: false
                      }}
                      editable={false}
                      theme={tokyoNight}
                    />
                  )}
                </pre>
              )}
              {!response && (
                <div className="p-2 md:h-full">
                  <EmptyState secondaryText="Enter a flag key, entity ID, and optional context to evaluate" />
                </div>
              )}
            </div>
          </>
        )}
        {flags.length === 0 && (
          <div className="mt-12 w-full">
            <EmptyState
              text="Create Flag"
              secondaryText="At least one flag must exist to use the console"
              onClick={() => navigate(`/namespaces/${namespace.key}/flags/new`)}
            />
          </div>
        )}
      </div>
    </>
  );
}
