import { Form, Formik, useFormikContext } from 'formik';
import hljs from 'highlight.js';
import javascript from 'highlight.js/lib/languages/json';
import 'highlight.js/styles/tokyo-night-dark.css';
import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import * as Yup from 'yup';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/Button';
import Combobox, { ISelectable } from '~/components/forms/Combobox';
import Input from '~/components/forms/Input';
import TextArea from '~/components/forms/TextArea';
import { evaluate, listFlags } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import {
  jsonValidation,
  keyValidation,
  requiredValidation
} from '~/data/validations';
import { IConsole } from '~/types/Console';
import { IFlag, IFlagList } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';

hljs.registerLanguage('json', javascript);

type SelectableFlag = IFlag & ISelectable;

function ResetOnNamespaceChange({ namespace }: { namespace: INamespace }) {
  const { resetForm } = useFormikContext();

  useEffect(() => {
    resetForm();
  }, [namespace, resetForm]);

  return null;
}

export default function Console() {
  const [flags, setFlags] = useState<SelectableFlag[]>([]);
  const [selectedFlag, setSelectedFlag] = useState<SelectableFlag | null>(null);
  const [response, setResponse] = useState<string | null>(null);

  const { setError, clearError } = useError();
  const navigate = useNavigate();

  const { currentNamespace } = useNamespace();

  const loadData = useCallback(async () => {
    const initialFlagList = (await listFlags(
      currentNamespace.key
    )) as IFlagList;
    const { flags } = initialFlagList;

    setFlags(
      flags.map((flag) => {
        const status = flag.enabled ? 'active' : 'inactive';

        return {
          ...flag,
          status,
          filterValue: flag.key,
          displayValue: flag.name
        };
      })
    );
  }, [currentNamespace.key]);

  const handleSubmit = (values: IConsole) => {
    const { flagKey, entityId, context } = values;
    // need to unescape the context string
    const parsed = context ? JSON.parse(context) : undefined;

    const rest = {
      entityId,
      context: parsed
    };

    evaluate(currentNamespace.key, flagKey, rest)
      .then((resp) => {
        setResponse(JSON.stringify(resp, null, 2));
      })
      .catch((err) => {
        setResponse(err.message);
      });
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

  const initialvalues: IConsole = {
    flagKey: selectedFlag?.key || '',
    entityId: uuidv4(),
    context: undefined
  };

  return (
    <>
      <div className="flex flex-col">
        <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl">
          Console
        </h1>
        <p className="mt-2 text-sm text-gray-500">
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
                  handleSubmit(values);
                }}
                onReset={() => {
                  setResponse(null);
                  setSelectedFlag(null);
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
                          <Combobox<SelectableFlag>
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
                          <Input
                            className="mt-1"
                            name="entityId"
                            id="entityId"
                            type="text"
                          />
                        </div>
                        <div className="col-span-3">
                          <label
                            htmlFor="context"
                            className="block text-sm font-medium text-gray-700"
                          >
                            Request Context
                          </label>
                          <TextArea
                            rows={10}
                            name="context"
                            id="context"
                            className="mt-1"
                            placeholder="{}"
                          />
                        </div>
                      </div>
                      <div className="flex justify-end">
                        <Button
                          type="reset"
                          onClick={(e) => {
                            e.preventDefault();
                            formik.resetForm();
                            formik.setFieldValue('entityId', uuidv4());
                            formik.setFieldValue('context', '');
                          }}
                        >
                          Reset
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
                    <ResetOnNamespaceChange namespace={currentNamespace} />
                  </Form>
                )}
              </Formik>
            </div>
            <div className="mt-8 w-full overflow-hidden md:w-1/2 md:pl-4">
              {response && (
                <pre className="p-2 text-sm md:h-full">
                  <code className="json rounded-sm md:h-full">
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
              onClick={() =>
                navigate(`/namespaces/${currentNamespace.key}/flags/new`)
              }
            />
          </div>
        )}
      </div>
    </>
  );
}
