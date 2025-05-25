import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import * as Yup from 'yup';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  useCreateFlagMutation,
  useUpdateFlagMutation
} from '~/app/flags/flagsApi';
import Rollouts from '~/app/flags/rollouts/Rollouts';
import Rules from '~/app/flags/rules/Rules';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import { UnsavedChangesModalWrapper } from '~/components/UnsavedChangesModal';
import Input from '~/components/forms/Input';
import Toggle from '~/components/forms/Toggle';
import Variants from '~/components/variants/Variants';

import { IDistribution } from '~/types/Distribution';
import { FlagType, IFlag } from '~/types/Flag';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { cls, getRevision, stringAsKey } from '~/utils/helpers';

import { FlagFormProvider } from './FlagFormContext';
import { MetadataForm } from './MetadataForm';
import MetadataFormErrorBoundary from './MetadataFormErrorBoundary';

// Form-specific interface that allows null type during initialization
export interface IFlagFormValues extends Omit<IFlag, 'type'> {
  type: FlagType | null;
}

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
      "Can be either 'true' or 'false'. The default value is returned when no rollouts match."
  }
];

const flagValidationSchema = Yup.object({
  type: Yup.string().required('Please select a flag type'),
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
  { name: 'Rules', id: 'rules' }
];

const booleanFlagTabs = [{ name: 'Rollouts', id: 'rollouts' }];

function FlagTypeSelector({
  selectedType,
  onTypeSelect
}: {
  selectedType: FlagType | null;
  onTypeSelect: (type: FlagType) => void;
}) {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100">
          Choose Flag Type
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Select the type of flag you want to create
        </p>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {flagTypes.map((flagType) => (
          <div
            data-testid={flagType.id}
            key={flagType.id}
            onClick={() => onTypeSelect(flagType.id)}
            className={cls(
              'relative flex cursor-pointer flex-col rounded-lg border p-4 shadow-sm focus:outline-none hover:border-ring bg-secondary/20 dark:bg-secondary/80',
              {
                'border-ring ring-ring/50 ring-1 shadow-xs':
                  selectedType === flagType.id
              }
            )}
          >
            <div className="flex flex-1">
              <div className="flex flex-col">
                <div className="flex items-center">
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {flagType.name}
                  </span>
                </div>
                <p className="mt-2 flex items-center text-sm text-muted-foreground">
                  {flagType.description}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

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
        environmentKey: environment.key,
        namespaceKey: namespace.key,
        values: v,
        revision
      }).unwrap();
    }

    return updateFlag({
      environmentKey: environment.key,
      namespaceKey: namespace.key,
      flagKey: flag?.key,
      values: v,
      revision
    }).unwrap();
  };

  const initialValues: IFlagFormValues = {
    key: flag?.key || '',
    name: flag?.name || '',
    description: flag?.description || '',
    type: flag ? flag.type : null,
    enabled: flag?.enabled || false,
    variants: flag?.variants || [],
    rules: flag?.rules || [],
    rollouts: flag?.rollouts || [],
    defaultVariant: flag?.defaultVariant,
    metadata: flag?.metadata || {}
  };

  const [hasMetadataErrors, setHasMetadataErrors] = useState(false);

  const [selectedTab, setSelectedTab] = useState(
    flag?.type == FlagType.VARIANT ? 'variants' : 'rollouts'
  );

  const tabs =
    flag?.type === FlagType.VARIANT ? variantFlagTabs : booleanFlagTabs;

  return (
    <Formik<IFlagFormValues>
      enableReinitialize
      initialValues={initialValues}
      onSubmit={(values, { setSubmitting }) => {
        // Assert type is not null since validation ensures this
        const flagValues = values as IFlag;

        // Assign ranks based on array indices before submission
        if (flagValues.rules) {
          flagValues.rules = flagValues.rules.map((rule, index) => ({
            ...rule,
            rank: index
          }));
        }

        if (flagValues.rollouts) {
          flagValues.rollouts = flagValues.rollouts.map((rollout, index) => ({
            ...rollout,
            rank: index
          }));
        }

        handleSubmit(flagValues)
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
      validate={(values: IFlagFormValues) => {
        let errors: any = {};

        if (isNew && !values.type) {
          errors.type = 'Please select a flag type';
          return errors;
        }

        values.rules?.forEach((rule, index) => {
          var ruleErrors: any = {};
          if (!validRollout(rule.distributions)) {
            ruleErrors.rollouts = 'Rollouts must add up to 100%';
          }
          if (!rule.segments || rule.segments.length <= 0) {
            ruleErrors.segments = 'Segments length must be greater than 0';
          }
          if (ruleErrors.rollouts || ruleErrors.segments) {
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
          !formik.isSubmitting &&
          (isNew ? formik.values.type !== null : true)
        );

        const form = (
          <Form className="space-y-6 p-1 sm:overflow-hidden sm:rounded-md">
            {isNew && (
              <FlagTypeSelector
                selectedType={formik.values.type}
                onTypeSelect={(type) => {
                  formik.setFieldValue('type', type);
                  formik.setFieldValue('enabled', false);
                }}
              />
            )}

            {(!isNew || formik.values.type) && (
              <div className="space-y-6">
                <div className="grid grid-cols-3 gap-6">
                  {formik.values.type === FlagType.VARIANT && (
                    <div className="col-span-3 md:col-span-2">
                      <div className="flex items-center justify-between">
                        <div>
                          <label
                            htmlFor="enabled"
                            className="block text-sm font-medium"
                          >
                            Enabled
                          </label>
                          <p className="text-sm text-muted-foreground">
                            Allows the flag to be evaluated
                          </p>
                        </div>
                        <Toggle
                          id="enabled"
                          name="enabled"
                          checked={enabled}
                          onChange={(e) => {
                            formik.setFieldValue('enabled', e);
                          }}
                        />
                      </div>
                    </div>
                  )}
                  {formik.values.type === FlagType.BOOLEAN && (
                    <div className="col-span-3 md:col-span-2">
                      <div className="flex items-center justify-between">
                        <div>
                          <label
                            htmlFor="defaultValue"
                            className="block text-sm font-medium text-gray-700 dark:text-gray-200"
                          >
                            Default Value
                          </label>
                          <p className="text-sm text-muted-foreground">
                            The default value returned when no rollouts match
                          </p>
                        </div>
                        <Toggle
                          id="defaultValue"
                          name="defaultValue"
                          checked={enabled}
                          onChange={(e) => {
                            formik.setFieldValue('enabled', e);
                          }}
                        />
                      </div>
                    </div>
                  )}
                  <div className="col-span-2">
                    <label
                      htmlFor="name"
                      className="block text-sm font-medium text-gray-700 dark:text-gray-200"
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
                  {isNew && (
                    <div className="col-span-2">
                      <label
                        htmlFor="key"
                        className="block text-sm font-medium text-gray-700 dark:text-gray-200"
                      >
                        Key
                      </label>
                      <Input
                        className="mt-1"
                        name="key"
                        id="key"
                        onChange={(e) => {
                          const formatted = stringAsKey(e.target.value);
                          formik.setFieldValue('key', formatted);
                        }}
                      />
                    </div>
                  )}
                  <div className="col-span-3">
                    <div className="flex justify-between">
                      <label
                        htmlFor="description"
                        className="block text-sm font-medium text-gray-700 dark:text-gray-200"
                      >
                        Description
                      </label>
                      <span
                        className="text-xs text-gray-500 dark:text-gray-400"
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
                        className="block text-sm font-medium text-gray-700 dark:text-gray-200"
                      >
                        Metadata
                      </label>
                      <span
                        className="text-xs text-gray-500 dark:text-gray-400"
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
              </div>
            )}

            {/* Tabs Section */}
            {flag && (
              <div>
                <div className="flex flex-row">
                  <div className="border-b-2 border-gray-200 dark:border-gray-700">
                    <nav className="-mb-px flex space-x-8">
                      {tabs.map((tab) => (
                        <div
                          role="link"
                          key={tab.name}
                          onClick={(e) => {
                            e.preventDefault();
                            setSelectedTab(tab.id!);
                          }}
                          className={cls(
                            'cursor-pointer whitespace-nowrap border-b-2 px-1 py-2 font-medium',
                            {
                              'border-violet-500 text-violet-600 dark:text-violet-400':
                                tab.id === selectedTab,
                              'border-transparent text-gray-600 hover:border-gray-300 hover:text-gray-700 dark:text-gray-300 dark:hover:border-gray-500 dark:hover:text-gray-200':
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
                {selectedTab == 'variants' && <Variants variants={variants!} />}
                {selectedTab == 'rollouts' && (
                  <Rollouts flag={flag} rollouts={rollouts!} />
                )}
                {selectedTab == 'rules' && <Rules flag={flag} rules={rules!} />}
              </div>
            )}

            <div className="flex justify-end">
              <Button type="button" onClick={() => navigate(-1)}>
                Cancel
              </Button>
              {(!isNew || formik.values.type) && (
                <div className="relative inline-block">
                  <Button
                    variant="primary"
                    className="ml-3 min-w-[80px]"
                    type="submit"
                    disabled={disableSave || hasMetadataErrors}
                  >
                    {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                  </Button>
                  {formik.dirty && formik.isValid && (
                    <div className="absolute -right-1 -top-1 h-3 w-3">
                      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-violet-100 opacity-75 dark:bg-violet-700"></span>
                    </div>
                  )}
                </div>
              )}
            </div>
          </Form>
        );
        return (
          <FlagFormProvider formik={formik}>
            {isNew ? (
              form
            ) : (
              <UnsavedChangesModalWrapper formik={formik}>
                {form}
              </UnsavedChangesModalWrapper>
            )}
          </FlagFormProvider>
        );
      }}
    </Formik>
  );
}
