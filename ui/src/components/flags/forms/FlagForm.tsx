import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/react/20/solid';
import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import * as Yup from 'yup';
import {
  useCreateFlagMutation,
  useUpdateFlagMutation
} from '~/app/flags/flagsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { Button } from '~/components/Button';
import Input from '~/components/forms/Input';
import Toggle from '~/components/forms/Toggle';
import Loading from '~/components/Loading';
import { MetadataForm } from '~/components/flags/forms/MetadataForm';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { FlagType, IFlag, IFlagBase } from '~/types/Flag';
import { cls, copyTextToClipboard, getRevision, stringAsKey } from '~/utils/helpers';
import MetadataFormErrorBoundary from './MetadataFormErrorBoundary';
import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';

const flagTypes = [
  {
    id: FlagType.VARIANT,
    name: 'Variant',
    description:
      "Can have multiple string values or 'variants'. Rules can be used to determine which variant is returned."
  },
  {
    id: FlagType.BOOLEAN,
    name: 'Boolean',
    description:
      "Can be either 'true' or 'false'. Rollouts can be used to determine which value is returned."
  }
];

const flagValidationSchema = Yup.object({
  key: keyValidation,
  name: requiredValidation,
  metadata: Yup.object()
});

export default function FlagForm(props: { flag?: IFlag }) {
  const { flag } = props;

  const isNew = flag === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';

  const navigate = useNavigate();

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const revision = getRevision();

  const [createFlag] = useCreateFlagMutation();
  const [updateFlag] = useUpdateFlagMutation();

  const handleSubmit = (values: IFlagBase) => {
    if (isNew) {
      return createFlag({
        environmentKey: environment.name,
        namespaceKey: namespace.key,
        values,
        revision
      }).unwrap();
    }

    return updateFlag({
      environmentKey: environment.name,
      namespaceKey: namespace.key,
      flagKey: flag?.key,
      values,
      revision
    }).unwrap();
  };

  const initialValues: IFlagBase = {
    key: flag?.key || '',
    name: flag?.name || '',
    description: flag?.description || '',
    type: flag?.type || FlagType.VARIANT,
    enabled: flag?.enabled || false,
    defaultVariant: flag?.defaultVariant,
    metadata: flag?.metadata || {}
  };

  const [keyCopied, setKeyCopied] = useState(false);
  const [hasMetadataErrors, setHasMetadataErrors] = useState(false);

  return (
    <Formik
      enableReinitialize
      initialValues={initialValues}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess(
              `Successfully ${submitPhrase.toLocaleLowerCase()}d flag`
            );
            if (isNew) {
              navigate(`/namespaces/${namespace.key}/flags/${values.key}`);
            }
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
      validationSchema={flagValidationSchema}
    >
      {(formik) => {
        const { enabled } = formik.values;
        return (
          <Form className="px-1 sm:overflow-hidden sm:rounded-md">
            <div className="space-y-6">
              <div className="grid grid-cols-3 gap-6">
                {formik.values.type === FlagType.VARIANT && (
                  <div className="col-span-3 md:col-span-2">
                    <Toggle
                      id="enabled"
                      name="enabled"
                      label="Enabled"
                      checked={enabled}
                      onChange={(e) => {
                        formik.setFieldValue('enabled', e);
                      }}
                    />
                  </div>
                )}
                <div className="col-span-2">
                  <label
                    htmlFor="name"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Name
                  </label>
                  <Input
                    className="mt-1"
                    name="name"
                    id="name"
                    autoFocus={isNew}
                    onChange={(e) => {
                      // check if the name and key are currently in sync
                      // we do this so we don't override a custom key value
                      if (
                        isNew &&
                        (formik.values.key === '' ||
                          formik.values.key === stringAsKey(formik.values.name))
                      ) {
                        formik.setFieldValue(
                          'key',
                          stringAsKey(e.target.value)
                        );
                      }
                      formik.handleChange(e);
                    }}
                  />
                </div>
                <div className="col-span-2">
                  <label
                    htmlFor="key"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Key
                  </label>
                  <div
                    className={cls({
                      'flex items-center justify-between': !isNew
                    })}
                  >
                    <Input
                      className={cls('mt-1', { 'md:mr-2': !isNew })}
                      name="key"
                      id="key"
                      disabled={!isNew}
                      onChange={(e) => {
                        const formatted = stringAsKey(e.target.value);
                        formik.setFieldValue('key', formatted);
                      }}
                    />
                    {!isNew && (
                      <button
                        aria-label="Copy to clipboard"
                        title="Copy to Clipboard"
                        className="hidden md:block"
                        onClick={(e) => {
                          e.preventDefault();
                          copyTextToClipboard(flag?.key || '');
                          setKeyCopied(true);
                          setTimeout(() => {
                            setKeyCopied(false);
                          }, 2000);
                        }}
                      >
                        <CheckIcon
                          className={cls(
                            'invisible absolute m-auto h-5 w-5 justify-center align-middle text-green-400 opacity-0 transition-opacity duration-300 ease-in-out',
                            {
                              'visible opacity-100': keyCopied
                            }
                          )}
                        />
                        <ClipboardDocumentIcon
                          className={cls(
                            'visible m-auto h-5 w-5 justify-center align-middle text-gray-300 opacity-100 transition-opacity duration-300 ease-in-out hover:text-gray-400',
                            {
                              'invisible opacity-0': keyCopied
                            }
                          )}
                        />
                      </button>
                    )}
                  </div>
                </div>
                <div className="col-span-3">
                  <label
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Type
                  </label>
                  <fieldset className="mt-2">
                    <legend className="sr-only">Type</legend>
                    <div className="space-y-5">
                      {flagTypes.map((flagType) => (
                        <div
                          key={flagType.id}
                          className="relative flex items-start"
                        >
                          <div className="flex h-5 items-center">
                            <input
                              id={flagType.id}
                              aria-describedby={`${flagType.id}-description`}
                              name="type"
                              type="radio"
                              disabled={!isNew}
                              className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                              onChange={() => {
                                formik.setFieldValue('type', flagType.id);
                                formik.setFieldValue('enabled', false);
                              }}
                              checked={flagType.id === formik.values.type}
                              value={flagType.id}
                            />
                          </div>
                          <div className="ml-3 text-sm">
                            <label
                              htmlFor={flagType.id}
                              className="font-medium text-gray-700"
                            >
                              {flagType.name}
                            </label>
                            <p
                              id={`${flagType.id}-description`}
                              className="text-gray-500"
                            >
                              {flagType.description}
                            </p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </fieldset>
                </div>
                <div className="col-span-3">
                  <div className="flex justify-between">
                    <label
                      htmlFor="description"
                      className="block text-sm font-medium text-gray-700"
                    >
                      Description
                    </label>
                    <span
                      className="text-xs text-gray-500"
                      id="description-optional"
                    >
                      Optional
                    </span>
                  </div>
                  <Input className="mt-1" name="description" id="description" />
                </div>
                <div className="col-span-3">
                  <div className="flex justify-between">
                    <label
                      htmlFor="metadata"
                      className="block text-sm font-medium text-gray-700"
                    >
                      Metadata
                    </label>
                    <span
                      className="text-xs text-gray-500"
                      id="metadata-optional"
                    >
                      Optional
                    </span>
                  </div>
                  <div className="mt-1">
                    <MetadataFormErrorBoundary
                      metadata={formik.values.metadata}
                      onChange={(metadata) =>
                        formik.setFieldValue('metadata', metadata)
                      }
                    >
                      <MetadataForm
                        metadata={formik.values.metadata}
                        onChange={(metadata) =>
                          formik.setFieldValue('metadata', metadata)
                        }
                        onErrorChange={setHasMetadataErrors}
                      />
                    </MetadataFormErrorBoundary>
                  </div>
                </div>
              </div>
              <div className="flex justify-end">
                <Button type="button" onClick={() => navigate(-1)}>
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  className="ml-3 min-w-[80px]"
                  type="submit"
                  disabled={
                    !(formik.dirty && formik.isValid && !formik.isSubmitting) ||
                    hasMetadataErrors
                  }
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                </Button>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
