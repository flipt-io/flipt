import { Field, FieldArray, useFormikContext } from 'formik';
import { useCallback, useMemo, useState } from 'react';

import { TextButton } from '~/components/Button';
import SegmentsPicker from '~/components/forms/SegmentsPicker';

import { DistributionType } from '~/types/Distribution';
import { IDistribution } from '~/types/Distribution';
import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { ISegment, segmentOperators } from '~/types/Segment';
import { FilterableVariant } from '~/types/Variant';

import { createSegmentHandlers } from '~/utils/formik-helpers';
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

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (ruleType !== DistributionType.Single) return null;

      if (distributions.length !== 1) {
        return null;
      }
      let selected = flag.variants?.find(
        (v) => v.key === distributions[0].variant
      );
      if (selected) {
        return {
          ...selected,
          displayValue: selected.name,
          filterValue: selected.key
        };
      }
      return null;
    });

  const ruleSegmentKeys = useMemo(() => rule.segments || [], [rule.segments]);

  const ruleSegments = useMemo(() => {
    return ruleSegmentKeys.map((s) => {
      const segment = segments.find((seg) => seg.key === s);
      if (!segment) {
        throw new Error(`Segment ${s} not found in segments`);
      }
      return {
        ...segment,
        displayValue: segment.name,
        filterValue: segment.key
      };
    });
  }, [ruleSegmentKeys, segments]);

  const fieldPrefix = `rules.[${rule.rank}].`;

  const formik = useFormikContext<IFlag>();

  // Initialize the rule with distributions if it doesn't have any
  const initializeDistributions = useCallback(() => {
    if (!rule.distributions) {
      formik.setFieldValue(`${fieldPrefix}distributions`, []);
    }
  }, [formik, fieldPrefix, rule.distributions]);

  // Use the shared segment handlers with the rule-specific update function
  const { handleSegmentAdd, handleSegmentRemove, handleSegmentReplace } =
    createSegmentHandlers<IRule, IFlag>(
      formik,
      rule,
      'rules',
      (original, segments) => ({
        ...original,
        segments
      }),
      initializeDistributions
    );

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
            <div>
              <label
                htmlFor={fieldPrefix + 'segments'}
                className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
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
              <div className="mt-6 flex space-x-8">
                {ruleSegments &&
                  ruleSegments.length > 1 &&
                  segmentOperators.map((segmentOperator, index) => (
                    <div className="flex space-x-2" key={index}>
                      <div>
                        <input
                          id={segmentOperator.id}
                          name={fieldPrefix + 'segmentOperator'}
                          type="radio"
                          className={cls(
                            'h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400'
                          )}
                          onChange={() => {
                            formik.setFieldValue(
                              fieldPrefix + 'segmentOperator',
                              segmentOperator.id
                            );
                          }}
                          checked={segmentOperator.id === rule.segmentOperator}
                          value={segmentOperator.id}
                        />
                      </div>
                      <div>
                        <label
                          htmlFor={segmentOperator.id}
                          className="block text-sm text-gray-700"
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
                className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
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
                            className="font-medium text-gray-700"
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
                id={fieldPrefix + 'distributions.[0].variant'}
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
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
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
                                  className="block truncate text-right text-sm text-gray-600 sm:mt-px sm:pr-2 sm:pt-2"
                                >
                                  {dist.variant}
                                </label>
                              </div>
                              <div className="relative sm:col-span-1">
                                <Field
                                  key={index}
                                  type="number"
                                  className={cls(
                                    'block w-full rounded-md border-gray-300 bg-gray-50 pl-7 pr-12 text-gray-900 shadow-xs focus:border-violet-300 focus:ring-violet-300 sm:text-sm'
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
              {formik.errors.rules && formik.errors.rules[rule.rank] && (
                <p className="mt-1 px-4 text-center text-sm text-gray-500 sm:px-6 sm:py-5">
                  Multi-variate rules must have distributions that add up to
                  100% or less.
                </p>
              )}
            </div>
          )}
        </div>
      </div>
      <div className="shrink-0 py-1">
        <div className="flex justify-end space-x-3">
          <TextButton
            disabled={formik.isSubmitting || !formik.dirty}
            onClick={() => formik.resetForm()}
          >
            Reset
          </TextButton>
        </div>
      </div>
    </div>
  );
}
