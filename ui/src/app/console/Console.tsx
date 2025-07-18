import { json } from '@codemirror/lang-json';
import { tokyoNight } from '@uiw/codemirror-theme-tokyo-night';
import CodeMirror from '@uiw/react-codemirror';
import { Form, Formik } from 'formik';
import { CodeIcon, RefreshCwIcon, SquareTerminalIcon } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router';
import { v4 as uuidv4 } from 'uuid';
import * as Yup from 'yup';

import { useListAuthProvidersQuery } from '~/app/auth/authApi';
import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { useListFlagsQuery } from '~/app/flags/flagsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { Button } from '~/components/Button';
import Combobox from '~/components/Combobox';
import Dropdown from '~/components/Dropdown';
import { PageHeader } from '~/components/Page';
import Well from '~/components/Well';
import Input from '~/components/forms/Input';
import { JsonEditor } from '~/components/json/JsonEditor';

import { IAuthMethod } from '~/types/Auth';
import { Command } from '~/types/Cli';
import { FilterableFlag, FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';

import { evaluate, evaluateURL } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import {
  contextValidation,
  keyValidation,
  requiredValidation
} from '~/data/validations';
import {
  copyTextToClipboard,
  generateCliCommand,
  generateCurlCommand,
  getErrorMessage
} from '~/utils/helpers';

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
  const { flagKey } = useParams();

  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);

  const {
    data: flagsData,
    error,
    isLoading: flagsLoading
  } = useListFlagsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  useEffect(() => {
    if (error) {
      setError(error);
      return;
    }
    clearError();
  }, [clearError, error, setError]);

  const flags = useMemo(() => {
    const initialFlags = flagsData?.flags || [];
    return initialFlags.map((flag) => {
      const status =
        flag.enabled || flag.type === FlagType.BOOLEAN ? 'active' : 'inactive';

      return {
        ...flag,
        status: status as 'active' | 'inactive',
        displayValue: `${flag.name} | ${flagTypeToLabel(flag.type)}`
      };
    });
  }, [flagsData]);

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

    evaluate(environment.key, namespace.key, flag.key, flag.type, rest)
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

    const whenNotDefault = (flag: string, value: string) =>
      value !== 'default' ? [{ key: flag, value: value }] : [];

    const contextOptions = Object.entries(parsed).map(([key, value]) => ({
      key: '--context',
      value: `${key}=${value}`
    }));

    const command = generateCliCommand({
      commandName: Command.Evaluate,
      arguments: [values.flagKey],
      options: [
        { key: '--entity-id', value: values.entityId },
        ...whenNotDefault('--namespace', namespace.key),
        ...whenNotDefault('--environment', environment.key),
        ...contextOptions
      ]
    });
    copyTextToClipboard(command)
      .then(() => {
        setSuccess('Command copied to clipboard');
      })
      .catch((err: unknown) => {
        setError((err as Error).message);
      });
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
        namespaceKey: namespace.key,
        environmentKey: environment.key
      },
      headers,
      uri
    });

    copyTextToClipboard(command)
      .then(() => {
        setSuccess('Command copied to clipboard');
      })
      .catch((err: unknown) => {
        setError((err as Error).message);
      });
  };

  const initialvalues: ConsoleFormValues = {
    flagKey: selectedFlag?.key || '',
    entityId: uuidv4(),
    context: '{}'
  };

  // Add a ref to hold Formik helpers
  const formikHelpersRef = useRef<any>(null);

  useEffect(() => {
    if (!formikHelpersRef.current) return;
    const formik = formikHelpersRef.current;

    if (flagKey && flags.length > 0) {
      const found = flags.find((f) => f.key === flagKey);
      if (found && (!selectedFlag || selectedFlag.key !== found.key)) {
        setSelectedFlag(found);
        formik.setFieldValue('flagKey', found.key);
      }
    } else if (!flagKey) {
      setSelectedFlag(null);
      formik.setFieldValue('flagKey', '');
    }
  }, [flagKey, flags, selectedFlag]);

  return (
    <>
      <PageHeader title="Playground" />
      <p className="mt-2 text-sm text-muted-foreground">
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
                {(formik) => {
                  formikHelpersRef.current = formik;

                  return (
                    <Form className="px-1 sm:overflow-hidden sm:rounded-md">
                      <div className="space-y-6">
                        <div className="grid grid-cols-3 gap-6">
                          <div className="col-span-3">
                            <label
                              htmlFor="flagKey"
                              className="block text-sm font-medium text-gray-700 dark:text-gray-200"
                            >
                              Flag
                            </label>
                            <Combobox<FilterableFlag>
                              id="flagKey"
                              name="flagKey"
                              className="mt-1 w-full"
                              placeholder="Select or search for a flag"
                              values={flags}
                              selected={selectedFlag}
                              setSelected={(flag) => {
                                setSelectedFlag(flag);
                                formik.setFieldValue(
                                  'flagKey',
                                  flag?.key || ''
                                );
                                if (flag) {
                                  navigate(
                                    `/namespaces/${namespace.key}/playground/${flag.key}`,
                                    { replace: true }
                                  );
                                } else {
                                  navigate(
                                    `/namespaces/${namespace.key}/playground`,
                                    { replace: true }
                                  );
                                }
                              }}
                              disabled={flagsLoading}
                            />
                          </div>
                          <div className="col-span-3">
                            <label
                              htmlFor="entityId"
                              className="block text-sm font-medium text-gray-700 dark:text-gray-200"
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
                                <RefreshCwIcon className="h-4 w-4 text-gray-400 dark:text-gray-300" />
                              </Button>
                            </div>
                          </div>
                          <div className="col-span-3">
                            <label
                              htmlFor="context"
                              className="block text-sm font-medium text-gray-700 dark:text-gray-200"
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
                    </Form>
                  );
                }}
              </Formik>
            </div>
            <div className="mt-10 w-full overflow-hidden md:w-1/2 md:pl-4">
              {response && (
                <pre className="rounded-md bg-[#1a1b26] p-2 text-sm md:h-full">
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
                  <Well>
                    <CodeIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
                    <h3 className="text-lg font-medium text-muted-foreground mb-2">
                      Ready to Evaluate
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      Enter a flag key, entity ID, and optional context to
                      evaluate
                    </p>
                  </Well>
                </div>
              )}
            </div>
          </>
        )}
        {flags.length === 0 && (
          <div className="mt-12 w-full">
            <Well>
              <CodeIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No Flags Available
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                At least one flag must exist to use the console
              </p>
              <Button
                aria-label="New Flag"
                variant="primary"
                onClick={() =>
                  navigate(`/namespaces/${namespace.key}/flags/new`)
                }
              >
                Create Your First Flag
              </Button>
            </Well>
          </div>
        )}
      </div>
    </>
  );
}
