import { ArrowPathIcon } from '@heroicons/react/20/solid';
import { Form, Formik, useFormikContext } from 'formik';
import hljs from 'highlight.js';
import javascript from 'highlight.js/lib/languages/json';
import 'highlight.js/styles/tomorrow-night-bright.css';
import React, { useCallback, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import * as Yup from 'yup';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { ContextEditor } from '~/components/console/ContextEditor';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Combobox from '~/components/forms/Combobox';
import Input from '~/components/forms/Input';
import { evaluateURL, evaluateV2, listFlags } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import {
  jsonValidation,
  keyValidation,
  requiredValidation
} from '~/data/validations';
import {
  FilterableFlag,
  FlagType,
  flagTypeToLabel,
  IFlag,
  IFlagList
} from '~/types/Flag';
import { INamespace } from '~/types/Namespace';
import {
  classNames,
  copyTextToClipboard,
  generateCurlCommand,
  getErrorMessage
} from '~/utils/helpers';

hljs.registerLanguage('json', javascript);

function ResetOnNamespaceChange({ namespace }: { namespace: INamespace }) {
  const { resetForm } = useFormikContext();

  useEffect(() => {
    resetForm();
  }, [namespace, resetForm]);

  return null;
}

interface ConsoleFormValues {
  flagKey: string;
  entityId: string;
  context: string;
}

export default function Console() {
  const [flags, setFlags] = useState<FilterableFlag[]>([]);
  const [selectedFlag, setSelectedFlag] = useState<FilterableFlag | null>(null);
  const [response, setResponse] = useState<string | null>(null);
  const [hasEvaluationError, setHasEvaluationError] = useState<boolean>(false);

  const { setError, clearError } = useError();
  const navigate = useNavigate();
  const {setSuccess} = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);

  const loadData = useCallback(async () => {
    const initialFlagList = (await listFlags(namespace.key)) as IFlagList;
    const { flags } = initialFlagList;

    setFlags(
      flags.map((flag) => {
        const status = flag.enabled ? 'active' : 'inactive';

        return {
          ...flag,
          status,
          filterValue: flag.key,
          displayValue: `${flag.name} | ${flagTypeToLabel(flag.type)}`
        };
      })
    );
  }, [namespace.key]);

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

    evaluateV2(namespace.key, flag.key, flag.type, rest)
      .then((resp) => {
        setHasEvaluationError(false);
        setResponse(JSON.stringify(resp, null, 2));
      })
      .catch((err) => {
        setHasEvaluationError(true);
        setResponse('error: ' + err.message);
      });
  };

  const handleCopyAsCurl = (values: ConsoleFormValues) => {
    let parsed = null;
    try {
      // need to unescape the context string
      parsed = JSON.parse(values.context);
    } catch (err) {
      setHasEvaluationError(true);
      setResponse('error: ' + getErrorMessage(err));
      return;
    }

    let route =
      selectedFlag?.type === FlagType.BOOLEAN ? '/boolean' : '/variant';
    const uri = window.location.origin + evaluateURL + route;

    const command = generateCurlCommand('POST', uri, {
      ...values,
      context: parsed,
      namespaceKey: namespace.key
    });

    copyTextToClipboard(command);

    setSuccess("Command copied to clipboard.")
  };

  useEffect(() => {
    hljs.highlightAll();
  }, [response]);

  useEffect(() => {
    loadData()
      .then(() => clearError())
      .catch((err) => {
        setError(err);
      });
  }, [clearError, loadData, setError]);

  const initialvalues: ConsoleFormValues = {
    flagKey: selectedFlag?.key || '',
    entityId: uuidv4(),
    context: '{}'
  };

  return (
    <>
      <div className="flex flex-col">
        <h1 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl">
          Console
        </h1>
        <p className="text-gray-500 mt-2 text-sm">
          See the results of your flag evaluations and debug any issues
        </p>
      </div>
      <div className="flex flex-col md:flex-row">
        {flags.length > 0 && (
          <>
            <div className="mt-8 w-full overflow-hidden md:w-1/2">
              <Formik
                initialValues={initialvalues}
                validationSchema={Yup.object({
                  flagKey: keyValidation,
                  entityId: requiredValidation,
                  context: jsonValidation
                })}
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
                            className="text-gray-700 block text-sm font-medium"
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
                            className="text-gray-700 block text-sm font-medium"
                          >
                            Entity ID
                          </label>
                          <div className="flex items-center justify-between">
                            <Input
                              className="mr-2 mt-1"
                              name="entityId"
                              id="entityId"
                              type="text"
                            />
                            <button
                              aria-label="New Entity ID"
                              title="New Entity ID"
                              className="hidden md:block"
                              onClick={(e) => {
                                e.preventDefault();
                                formik.setFieldValue('entityId', uuidv4());
                              }}
                            >
                              <ArrowPathIcon
                                className={classNames(
                                  'text-gray-400 m-auto h-5 w-5 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-gray-500'
                                )}
                              />
                            </button>
                          </div>
                        </div>
                        <div className="col-span-3">
                          <label
                            htmlFor="context"
                            className="text-gray-700 block text-sm font-medium"
                          >
                            Request Context
                          </label>
                          <div className="nightwind-prevent mt-1">
                            <ContextEditor
                              id="context"
                              setValue={(v) => {
                                formik.setFieldValue('context', v);
                              }}
                            />
                          </div>
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <Button
                          primary
                          className="ml-3"
                          type="button"
                          disabled={!(formik.dirty && formik.isValid)}
                          onClick={() => {
                            handleCopyAsCurl(formik.values);
                          }}
                        >
                          Copy as curl
                        </Button>
                        <Button
                          primary
                          className="ml-3"
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
            <div className="mt-8 w-full overflow-hidden md:w-1/2 md:pl-4">
              {response && (
                <pre className="p-2 text-sm md:h-full">
                  <code
                    className={classNames(
                      hasEvaluationError ? 'border-red-400 border-4' : '',
                      'json rounded-sm md:h-full'
                    )}
                  >
                    {response as React.ReactNode}
                  </code>
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
