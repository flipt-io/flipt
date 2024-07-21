import { Field, FieldArray, Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import {
  useUpdateDistributionMutation,
  useUpdateRuleMutation
} from '~/app/flags/rulesApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import TextButton from '~/components/forms/buttons/TextButton';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import Loading from '~/components/Loading';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { DistributionType } from '~/types/Distribution';
import { IEvaluatable, IVariantRollout } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { INamespace } from '~/types/Namespace';
import {
  FilterableSegment,
  ISegment,
  segmentOperators,
  SegmentOperatorType
} from '~/types/Segment';
import { FilterableVariant, toFilterableVariant } from '~/types/Variant';
import { cls } from '~/utils/helpers';
import { distTypes } from './RuleForm';
import SingleDistributionFormInput from '~/components/rules/forms/SingleDistributionForm';

type QuickEditRuleFormProps = {
  flag: IFlag;
  rule: IEvaluatable;
  segments: ISegment[];
  onSuccess?: () => void;
};

export const validRollout = (rollouts: IVariantRollout[]): boolean => {
  const sum = rollouts.reduce(function (acc, d) {
    return acc + Number(d.distribution.rollout);
  }, 0);

  return sum <= 100;
};

interface RuleFormValues {
  segmentKeys: FilterableSegment[];
  segmentKey?: string;
  rollouts: IVariantRollout[];
  operator: SegmentOperatorType;
}

export default function QuickEditRuleForm(props: QuickEditRuleFormProps) {
  const { onSuccess, flag, rule, segments } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace) as INamespace;

  const ruleType =
    rule.rollouts.length === 0
      ? DistributionType.None
      : rule.rollouts.length === 1
        ? DistributionType.Single
        : DistributionType.Multi;

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (ruleType !== DistributionType.Single) return null;

      if (rule.rollouts.length !== 1) {
        return null;
      }

      return toFilterableVariant(rule.rollouts[0].variant);
    });

  const readOnly = useSelector(selectReadonly);

  const [updateRule] = useUpdateRuleMutation();
  const [updateDistribution] = useUpdateDistributionMutation();

  const handleSubmit = async (values: RuleFormValues) => {
    const originalRuleSegments = rule.segments.map((s) => s.key);
    const comparableRuleSegments = values.segmentKeys.map((s) => s.key);

    const segmentsDidntChange =
      comparableRuleSegments.every((rs) => {
        return originalRuleSegments.includes(rs);
      }) && comparableRuleSegments.length === originalRuleSegments.length;

    if (!segmentsDidntChange || values.operator !== rule.operator) {
      // update segment if changed
      try {
        await updateRule({
          namespaceKey: namespace.key,
          flagKey: flag.key,
          ruleId: rule.id,
          values: {
            rank: rule.rank,
            segmentKeys: comparableRuleSegments,
            segmentOperator: values.operator
          }
        });
      } catch (err) {
        setError(err as Error);
        return;
      }
    }

    if (ruleType === DistributionType.Multi) {
      // update distributions that changed
      Promise.all(
        values.rollouts.map((rollout) => {
          const found = rule.rollouts.find(
            (r) => r.distribution.id === rollout.distribution.id
          );
          if (
            found &&
            found.distribution.rollout !== rollout.distribution.rollout
          ) {
            return updateDistribution({
              namespaceKey: namespace.key,
              flagKey: flag.key,
              ruleId: rule.id,
              distributionId: rollout.distribution.id,
              values: {
                variantId: rollout.distribution.variantId,
                rollout: rollout.distribution.rollout
              }
            });
          }
          return Promise.resolve();
        })
      );
    } else if (ruleType === DistributionType.Single && selectedVariant) {
      // update variant if changed
      if (rule.rollouts[0].distribution.variantId !== selectedVariant.id) {
        try {
          await updateDistribution({
            namespaceKey: namespace.key,
            flagKey: flag.key,
            ruleId: rule.id,
            distributionId: rule.rollouts[0].distribution.id,
            values: {
              variantId: selectedVariant.id,
              rollout: 100
            }
          });
        } catch (err) {
          setError(err as Error);
          return;
        }
      }
    }
  };

  return (
    <Formik
      enableReinitialize
      initialValues={{
        segmentKeys: rule.segments.map((segment) => ({
          ...segment,
          displayValue: segment.name,
          filterValue: segment.key
        })),
        // variantId: rule.rollouts[0].distribution.variantId,
        rollouts: rule.rollouts,
        operator: rule.operator
      }}
      validate={(values) => {
        const errors: any = {};
        if (!validRollout(values.rollouts)) {
          errors.rollouts = 'Rollouts must add up to 100%';
        }
        if (values.segmentKeys.length <= 0) {
          errors.segmentKeys = 'Segments length must be greater than 0';
        }
        return errors;
      }}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          ?.then(() => {
            clearError();
            setSuccess('Successfully updated rule');
            onSuccess && onSuccess();
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
    >
      {(formik) => {
        return (
          <Form className="bg-white flex h-full w-full flex-col overflow-y-scroll">
            <div className="w-full flex-1">
              <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                  <div>
                    <label
                      htmlFor="segmentKeys"
                      className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <div>
                      <FieldArray
                        name="segmentKeys"
                        render={(arrayHelpers) => (
                          <SegmentsPicker
                            readonly={readOnly}
                            segments={segments}
                            segmentAdd={(segment: FilterableSegment) =>
                              arrayHelpers.push(segment)
                            }
                            segmentRemove={(index: number) =>
                              arrayHelpers.remove(index)
                            }
                            segmentReplace={(
                              index: number,
                              segment: FilterableSegment
                            ) => arrayHelpers.replace(index, segment)}
                            selectedSegments={formik.values.segmentKeys}
                          />
                        )}
                      />
                    </div>
                    <div className="mt-6 flex space-x-8">
                      {formik.values.segmentKeys.length > 1 &&
                        segmentOperators.map((segmentOperator, index) => (
                          <div className="flex space-x-2" key={index}>
                            <div>
                              <input
                                id={segmentOperator.id}
                                name="operator"
                                type="radio"
                                className={cls(
                                  'text-violet-400 border-gray-300 h-4 w-4 focus:ring-violet-400',
                                  { 'cursor-not-allowed': readOnly }
                                )}
                                onChange={() => {
                                  formik.setFieldValue(
                                    'operator',
                                    segmentOperator.id
                                  );
                                }}
                                checked={
                                  segmentOperator.id === formik.values.operator
                                }
                                value={segmentOperator.id}
                                disabled={readOnly}
                                title={
                                  readOnly
                                    ? 'Not allowed in Read-Only mode'
                                    : undefined
                                }
                              />
                            </div>
                            <div>
                              <label
                                htmlFor={segmentOperator.id}
                                className="text-gray-700 block text-sm"
                              >
                                {segmentOperator.name}{' '}
                                <span className="font-light">
                                  {segmentOperator.meta}
                                </span>
                              </label>
                            </div>
                          </div>
                        ))}
                    </div>
                  </div>
                </div>
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                  <div>
                    <label
                      htmlFor="ruleType"
                      className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                    >
                      Type
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <fieldset>
                      <legend className="sr-only">Type</legend>
                      <div className="space-y-5">
                        {distTypes
                          .filter((dist) => dist.id === ruleType)
                          .map((dist) => (
                            <div
                              key={dist.id}
                              className="relative flex items-start"
                            >
                              <div className="text-sm">
                                <label
                                  htmlFor={dist.id}
                                  className="text-gray-700 font-medium"
                                >
                                  {dist.name}
                                </label>
                                <p
                                  id={`${dist.id}-description`}
                                  className="text-gray-500"
                                >
                                  {dist.description}
                                </p>
                              </div>
                            </div>
                          ))}
                      </div>
                    </fieldset>
                  </div>
                </div>

                {ruleType === DistributionType.Single &&
                  flag.variants &&
                  flag.variants.length > 0 && (
                    <SingleDistributionFormInput
                      id={`quick-variant-${rule.rank}`}
                      variants={flag.variants}
                      selectedVariant={selectedVariant}
                      setSelectedVariant={setSelectedVariant}
                    />
                  )}

                {ruleType === DistributionType.Multi && (
                  <div>
                    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                      <div>
                        <label
                          htmlFor="variantKey"
                          className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                        >
                          Variants
                        </label>
                      </div>
                    </div>
                    <FieldArray
                      name="rollouts"
                      render={() => (
                        <>
                          {formik.values.rollouts &&
                            formik.values.rollouts.length > 0 &&
                            formik.values.rollouts?.map((dist, index) => (
                              <div key={index}>
                                {dist && (
                                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-1">
                                    <div>
                                      <label
                                        htmlFor={`rollouts.[${index}].distribution.rollout`}
                                        className="text-gray-600 block truncate text-right text-sm sm:mt-px sm:pr-2 sm:pt-2"
                                      >
                                        {dist.variant.key}
                                      </label>
                                    </div>
                                    <div className="relative sm:col-span-1">
                                      <Field
                                        key={index}
                                        type="number"
                                        className={cls(
                                          'text-gray-900 bg-gray-50 border-gray-300 block w-full rounded-md pl-7 pr-12 shadow-sm focus:border-violet-300 focus:ring-violet-300 sm:text-sm',
                                          {
                                            'text-gray-500 bg-gray-100 cursor-not-allowed':
                                              readOnly
                                          }
                                        )}
                                        value={dist.distribution.rollout}
                                        name={`rollouts.[${index}].distribution.rollout`}
                                        // eslint-disable-next-line react/no-unknown-property
                                        typeof="number"
                                        step=".01"
                                        min="0"
                                        max="100"
                                        disabled={readOnly}
                                      />
                                      <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                        <span className="text-gray-500 sm:text-sm">
                                          %
                                        </span>
                                      </div>
                                    </div>
                                  </div>
                                )}
                              </div>
                            ))}
                        </>
                      )}
                    />
                    {formik.touched.rollouts && formik.errors.rollouts && (
                      <p className="text-gray-500 mt-1 px-4 text-center text-sm sm:px-6 sm:py-5">
                        Multi-variate rules must have distributions that add up
                        to 100% or less.
                      </p>
                    )}
                  </div>
                )}
              </div>
            </div>
            <div className="flex-shrink-0 py-1">
              <div className="flex justify-end space-x-3">
                <TextButton
                  disabled={formik.isSubmitting || !formik.dirty}
                  onClick={() => formik.resetForm()}
                >
                  Reset
                </TextButton>
                <TextButton
                  type="submit"
                  className="min-w-[80px]"
                  disabled={
                    !formik.isValid || formik.isSubmitting || !formik.dirty
                  }
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : 'Update'}
                </TextButton>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
