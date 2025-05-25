import { Field, FieldArray, useFormikContext } from 'formik';
import { useCallback, useEffect, useMemo, useState } from 'react';

import SegmentsPicker from '~/components/forms/SegmentsPicker';

import { DistributionType } from '~/types/Distribution';
import { IDistribution } from '~/types/Distribution';
import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { FilterableSegment, ISegment, segmentOperators } from '~/types/Segment';
import { FilterableVariant } from '~/types/Variant';

import { cls } from '~/utils/helpers';

import { distTypes } from './RuleForm';
import SingleDistributionFormInput from './SingleDistributionForm';

type QuickEditRuleFormProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  onSuccess?: () => void;
};

export const validRollout = (rollouts: IDistribution[]): boolean => {
  const sum = rollouts.reduce(function (acc, d) {
    return acc + Number(d.rollout);
  }, 0);

  return sum <= 100;
};

export default function QuickEditRuleForm(props: QuickEditRuleFormProps) {
  const { flag, rule, segments } = props;
  const formik = useFormikContext<IFlag>();

  // Ensure rule.distributions is always treated as an array
  const distributions = useMemo(
    () => rule.distributions || [],
    [rule.distributions]
  );

  const ruleType = useMemo(() => {
    return distributions.length === 0
      ? DistributionType.None
      : distributions.length === 1
        ? DistributionType.Single
        : DistributionType.Multi;
  }, [distributions.length]);

  // Calculate the actual index in the rules array by finding this rule's position
  const ruleIndex = useMemo(() => {
    const rules = formik.values.rules || [];
    // Try to find by ID first if available
    if (rule.id) {
      const index = rules.findIndex((r) => r.id === rule.id);
      if (index !== -1) return index;
    }
    // Fallback to position in array
    return rules.indexOf(rule);
  }, [formik.values.rules, rule]);

  // Use the calculated index for the field path
  const fieldPrefix = `rules.[${ruleIndex !== -1 ? ruleIndex : 0}].`;

  // Initialize selected variant based on current distribution
  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (ruleType !== DistributionType.Single || distributions.length !== 1) {
        return null;
      }

      const variantKey = distributions[0].variant;
      const variant = flag.variants?.find((v) => v.key === variantKey);

      if (!variant) {
        return null;
      }

      return {
        ...variant,
        displayValue: variant.name || variant.key
      };
    });

  // Handle variant selection
  const handleVariantChange = useCallback(
    (variant: FilterableVariant | null) => {
      // Update the local state
      setSelectedVariant(variant);

      // Update the formik state
      if (variant) {
        formik.setFieldValue(`${fieldPrefix}distributions`, [
          {
            variant: variant.key,
            rollout: 100
          }
        ]);
      } else {
        // Clear distributions if no variant is selected
        formik.setFieldValue(`${fieldPrefix}distributions`, []);
      }
    },
    [formik, fieldPrefix]
  );

  // Keep selected variant in sync with distributions
  useEffect(() => {
    if (ruleType === DistributionType.Single && distributions.length === 1) {
      const variantKey = distributions[0].variant;
      // Only update if the selected variant doesn't match the current distribution
      if (!selectedVariant || selectedVariant.key !== variantKey) {
        const variant = flag.variants?.find((v) => v.key === variantKey);
        if (variant) {
          setSelectedVariant({
            ...variant,
            displayValue: variant.name || variant.key
          });
        }
      }
    } else if (selectedVariant && ruleType !== DistributionType.Single) {
      // Clear selected variant if rule type is not single
      setSelectedVariant(null);
    }
  }, [distributions, flag.variants, ruleType, selectedVariant]);

  const ruleSegmentKeys = useMemo(() => rule.segments || [], [rule.segments]);

  const ruleSegments = useMemo(() => {
    return ruleSegmentKeys.map((s) => {
      const segment = segments.find((seg) => seg.key === s);
      if (!segment) {
        throw new Error(`Segment ${s} not found in segments`);
      }
      return {
        ...segment,
        displayValue: segment.name
      };
    });
  }, [ruleSegmentKeys, segments]);

  // Initialize the rule with distributions if it doesn't have any
  const initializeDistributions = useCallback(() => {
    if (!rule.distributions) {
      formik.setFieldValue(`${fieldPrefix}distributions`, []);
    }
  }, [formik, fieldPrefix, rule.distributions]);

  // Custom segment management functions
  const handleSegmentAdd = useCallback(
    (segment: FilterableSegment) => {
      // Double-check that we're working with the correct rule in formik values
      const rules = formik.values.rules || [];
      if (ruleIndex === -1 || ruleIndex >= rules.length) {
        return;
      }

      // Get current segments or initialize empty array
      const currentSegments = [...(rule.segments || [])];

      // Add the new segment key if it doesn't already exist
      if (!currentSegments.includes(segment.key)) {
        currentSegments.push(segment.key);

        // Initialize distributions if needed
        if (initializeDistributions) {
          initializeDistributions();
        }

        // Update the form value for this specific rule's segments
        formik.setFieldValue(`${fieldPrefix}segments`, currentSegments);
      }
    },
    [formik, fieldPrefix, rule, initializeDistributions, ruleIndex]
  );

  const handleSegmentRemove = useCallback(
    (index: number) => {
      // Double-check that we're working with the correct rule in formik values
      const rules = formik.values.rules || [];
      if (ruleIndex === -1 || ruleIndex >= rules.length) {
        return;
      }

      // Get current segments
      const currentSegments = [...(rule.segments || [])];

      // Remove the segment at the specified index
      if (index >= 0 && index < currentSegments.length) {
        currentSegments.splice(index, 1);

        // Update only the segments array
        formik.setFieldValue(`${fieldPrefix}segments`, currentSegments);
      }
    },
    [formik, fieldPrefix, rule, ruleIndex]
  );

  const handleSegmentReplace = useCallback(
    (index: number, segment: FilterableSegment) => {
      // Double-check that we're working with the correct rule in formik values
      const rules = formik.values.rules || [];
      if (ruleIndex === -1 || ruleIndex >= rules.length) {
        return;
      }

      // Get current segments
      const currentSegments = [...(rule.segments || [])];

      // Replace the segment at the specified index
      if (index >= 0 && index < currentSegments.length) {
        currentSegments[index] = segment.key;

        // Update only the segments array
        formik.setFieldValue(`${fieldPrefix}segments`, currentSegments);
      }
    },
    [formik, fieldPrefix, rule, ruleIndex]
  );

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
            <div>
              <label
                htmlFor={fieldPrefix + 'segments'}
                className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
              >
                Segment
              </label>
            </div>
            <div className="sm:col-span-2">
              <div>
                <FieldArray
                  name={fieldPrefix + 'segments'}
                  render={() => (
                    <SegmentsPicker
                      segments={segments}
                      segmentAdd={handleSegmentAdd}
                      segmentRemove={handleSegmentRemove}
                      segmentReplace={handleSegmentReplace}
                      selectedSegments={ruleSegments}
                    />
                  )}
                />
                {ruleSegments.length === 0 ? (
                  <div className="mt-1 text-sm text-red-500">
                    Segment is missing
                  </div>
                ) : null}
              </div>
              {ruleSegments && ruleSegments.length > 1 && (
                <div className="mt-4 flex space-x-8">
                  {segmentOperators.map((segmentOperator, index) => (
                    <div
                      className="flex space-x-2 cursor-pointer"
                      key={index}
                      onClick={() => {
                        formik.setFieldValue(
                          `${fieldPrefix}segmentOperator`,
                          segmentOperator.id
                        );
                      }}
                    >
                      <div>
                        <input
                          id={segmentOperator.id}
                          name={fieldPrefix + 'segmentOperator'}
                          type="radio"
                          className="h-4 w-4 border text-ring focus:ring-ring cursor-pointer"
                          checked={segmentOperator.id === rule.segmentOperator}
                          value={segmentOperator.id}
                          readOnly
                        />
                      </div>
                      <div>
                        <label
                          htmlFor={segmentOperator.id}
                          className="block text-sm text-gray-700 dark:text-gray-200 cursor-pointer"
                        >
                          {segmentOperator.name}{' '}
                          <span className="font-light dark:text-gray-300">
                            {segmentOperator.meta}
                          </span>
                        </label>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
          <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
            <div>
              <label
                htmlFor="ruleType"
                className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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
                      <div key={dist.id} className="relative flex items-start">
                        <div className="text-sm">
                          <label
                            htmlFor={dist.id}
                            className="font-medium text-gray-700 dark:text-gray-200"
                          >
                            {dist.name}
                          </label>
                          <p
                            id={`${dist.id}-description`}
                            className="text-gray-500 dark:text-gray-400"
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
                id={fieldPrefix + 'distributions.[0].variant'}
                variants={flag.variants}
                selectedVariant={selectedVariant}
                setSelectedVariant={handleVariantChange}
              />
            )}

          {ruleType === DistributionType.Multi && (
            <div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                <div>
                  <label
                    htmlFor="variantKey"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Variants
                  </label>
                </div>
              </div>
              <FieldArray
                name="rollouts"
                render={() => (
                  <>
                    {distributions.length > 0 &&
                      distributions.map((dist, index) => (
                        <div key={index}>
                          {dist && (
                            <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-1">
                              <div>
                                <label
                                  htmlFor={`${fieldPrefix}distributions.[${index}].rollout`}
                                  className="block truncate text-right text-sm text-gray-600 dark:text-gray-300 sm:mt-px sm:pr-2 sm:pt-2"
                                >
                                  {dist.variant}
                                </label>
                              </div>
                              <div className="relative sm:col-span-1">
                                <Field
                                  key={index}
                                  type="number"
                                  className={cls(
                                    'block w-full rounded-md border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 pl-7 pr-12 text-gray-900 dark:text-gray-100 shadow-xs focus:border-violet-300 focus:ring-violet-300 sm:text-sm'
                                  )}
                                  value={dist.rollout}
                                  name={`${fieldPrefix}distributions.[${index}].rollout`}
                                  // eslint-disable-next-line react/no-unknown-property
                                  typeof="number"
                                  step=".01"
                                  min="0"
                                  max="100"
                                />
                                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                                  <span className="text-gray-500 dark:text-gray-400 sm:text-sm">
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
              {formik.errors.rules && formik.errors.rules[ruleIndex] && (
                <p className="mt-1 px-4 text-center text-sm text-red-500 dark:text-red-400 sm:px-6 sm:py-5">
                  Multi-variate rules must have distributions that add up to
                  100% or less.
                </p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
