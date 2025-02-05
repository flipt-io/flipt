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
import { MetadataForm } from './MetadataForm';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { FlagType, IFlag, IFlagBase } from '~/types/Flag';
import {
  cls,
  copyTextToClipboard,
  getRevision,
  stringAsKey
} from '~/utils/helpers';
import MetadataFormErrorBoundary from './MetadataFormErrorBoundary';
import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { IDistribution } from '~/types/Distribution';
import { FlagFormProvider } from './FlagFormContext';
import Rollouts from '../rollouts/Rollouts';
import Variants from '../variants/Variants';
import Rules from '~/app/flags/rules/Rules';

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

export const validRollout = (rollouts: IDistribution[]): boolean => {
  const sum = rollouts.reduce(function (acc, d) {
    return acc + Number(d.rollout);
  }, 0);

  return sum <= 100;
};

const variantFlagTabs = [
  { name: 'Variants', id: 'variants' },
  { name: 'Rules', id: 'rules' },
  { name: 'Analytics', id: 'analytics' }
];

const booleanFlagTabs = [
  { name: 'Rollouts', id: 'rollouts' },
  { name: 'Analytics', id: 'analytics' }
];

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

  const handleSubmit = (values: IFlag) => {
    const v = {
      ...values,
      rollouts: values.type === FlagType.BOOLEAN ? values.rollouts : undefined,
      variants: values.type === FlagType.VARIANT ? values.variants : undefined,
      rules: values.type === FlagType.VARIANT ? values.rules : undefined,
      metadata: values.metadata
    };

    if (isNew) {
      return createFlag({
        environmentKey: environment.name,
        namespaceKey: namespace.key,
        values: v,
        revision
      }).unwrap();
    }

    return updateFlag({
      environmentKey: environment.name,
      namespaceKey: namespace.key,
      flagKey: flag?.key,
      values: v,
      revision
    }).unwrap();
  };

  const initialValues: IFlag = {
    key: flag?.key || '',
    name: flag?.name || '',
    description: flag?.description || '',
    type: flag?.type || FlagType.VARIANT,
    enabled: flag?.enabled || false,
    variants: flag?.variants || [],
    rules: flag?.rules || [],
    rollouts: flag?.rollouts || [],
    defaultVariant: flag?.defaultVariant,
    metadata: flag?.metadata || {}
  };

  const [keyCopied, setKeyCopied] = useState(false);
  const [hasMetadataErrors, setHasMetadataErrors] = useState(false);

  const [selectedTab, setSelectedTab] = useState(
    flag?.type == FlagType.VARIANT ? 'variants' : 'rollouts'
  );

  const tabs =
    flag?.type === FlagType.VARIANT ? variantFlagTabs : booleanFlagTabs;

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
      validate={(values: IFlag) => {
        let errors: any = {};
        values.rules?.forEach((rule, index) => {
          var ruleErrors: any = {};
          if (!validRollout(rule.distributions)) {
            ruleErrors.rollouts = 'Rollouts must add up to 100%';
          }
          if (!rule.segments || rule.segments.length <= 0) {
            ruleErrors.segmentKeys = 'Segments length must be greater than 0';
          }
          if (ruleErrors.rollouts || ruleErrors.segmentKeys) {
            if (!errors.rules) {
              errors.rules = [];
            }
            errors.rules[index] = ruleErrors;
          }
        });
        return errors;
      }}
      validationSchema={flagValidationSchema}
    >
      {(formik) => {
        const { enabled, variants, rollouts, rules } = formik.values;
        const disableSave = !(
          formik.dirty &&
          formik.isValid &&
          !formik.isSubmitting
        );
        return (
          <FlagFormProvider formik={formik}>
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
                            formik.values.key ===
                              stringAsKey(formik.values.name))
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
                    <Input
                      className="mt-1"
                      name="description"
                      id="description"
                    />
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

                {/* Tabs Section */}
                {flag && (
                  <>
                    <div className="mt-3 flex flex-row sm:mt-5">
                      <div className="border-b-2 border-gray-200">
                        <nav className="-mb-px flex space-x-8">
                          {tabs.map((tab) => (
                            <div
                              role="button"
                              key={tab.name}
                              onClick={() => setSelectedTab(tab.id!)}
                              className={cls(
                                'cursor-pointer whitespace-nowrap border-b-2 px-1 py-2 font-medium',
                                {
                                  'border-violet-500 text-violet-600':
                                    tab.id === selectedTab,
                                  'border-transparent text-gray-600 hover:border-gray-300 hover:text-gray-700':
                                    tab.id != selectedTab
                                }
                              )}
                            >
                              {tab.name}
                            </div>
                          ))}
                        </nav>
                      </div>
                    </div>
                    {selectedTab == 'variants' && (
                      <Variants flag={flag} variants={variants!} />
                    )}
                    {/* {selectedTab == 'rollouts' && (
                        <Rollouts flag={flag} rollouts={rollouts!} />
                      )}
                      {selectedTab == 'rules' && (
                        <Rules
                          flag={flag}
                          rules={rules!}
                          variants={variants!}
                        />
                      )} */}
                  </>
                )}
                <div className="flex justify-end">
                  <Button type="button" onClick={() => navigate(-1)}>
                    Cancel
                  </Button>
                  <Button
                    variant="primary"
                    className="ml-3 min-w-[80px]"
                    type="submit"
                    disabled={disableSave || hasMetadataErrors}
                  >
                    {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                  </Button>
                </div>
              </div>
            </Form>
          </FlagFormProvider>
        );
      }}
    </Formik>
  );
}
