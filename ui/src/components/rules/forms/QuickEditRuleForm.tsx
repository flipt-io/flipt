import { Field, FieldArray, useFormikContext } from 'formik';
import { useState } from 'react';
import SegmentsPicker from '~/components/forms/SegmentsPicker';
import { DistributionType } from '~/types/Distribution';
import { IRule } from '~/types/Rule';
import { IFlag } from '~/types/Flag';
import { FilterableSegment, ISegment, segmentOperators } from '~/types/Segment';
import { FilterableVariant, IVariant } from '~/types/Variant';
import { cls } from '~/utils/helpers';
import { distTypes } from './RuleForm';
import SingleDistributionFormInput from '~/components/rules/forms/SingleDistributionForm';
import { IDistribution } from '~/types/Distribution';

type QuickEditRuleFormProps = {
  flag: IFlag;
  rule: IRule;
  segments: ISegment[];
  variants: IVariant[];
  onSuccess?: () => void;
};

export const validRollout = (rollouts: IDistribution[]): boolean => {
  const sum = rollouts.reduce(function (acc, d) {
    return acc + Number(d.rollout);
  }, 0);

  return sum <= 100;
};

export default function QuickEditRuleForm(props: QuickEditRuleFormProps) {
  const { onSuccess, flag, rule, segments, variants } = props;

  const ruleType =
    rule.distributions.length === 0
      ? DistributionType.None
      : rule.distributions.length === 1
        ? DistributionType.Single
        : DistributionType.Multi;

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (ruleType !== DistributionType.Single) return null;

      if (rule.distributions.length !== 1) {
        return null;
      }
      let selected = variants.find(
        (v) => v.key === rule.distributions[0].variant
      );
      if (selected) {
        return {
          ...selected,
          displayValue: selected.name,
          filterValue: selected.id
        };
      }
      return null;
    });

  const ruleSegmentKeys = rule.segments || [];

  const ruleSegments = ruleSegmentKeys.map((s) => {
    const segment = props.segments.find((seg) => seg.key === s);
    if (!segment) {
      throw new Error(`Segment ${s} not found in segments`);
    }
    return {
      ...segment,
      displayValue: segment.name,
      filterValue: segment.key
    };
  });

  const fieldPrefix = `rules.[${rule.rank}].`;

  const formik = useFormikContext<IFlag>();

  return (
    <div className="flex h-full w-full flex-col">
      <div className="w-full flex-1">
        <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
          <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
            <div>
              <label
                htmlFor="segmentKeys"
                className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
              >
                Segment
              </label>
            </div>
            <div className="sm:col-span-2">
              <div>
                <FieldArray
                  name={fieldPrefix + 'segments'}
                  render={(arrayHelpers) => (
                    <SegmentsPicker
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
                    {rule.distributions &&
                      rule.distributions.length > 0 &&
                      rule.distributions?.map((dist, index) => (
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
                                    'block w-full rounded-md border-gray-300 bg-gray-50 pl-7 pr-12 text-gray-900 shadow-sm focus:border-violet-300 focus:ring-violet-300 sm:text-sm'
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
    </div>
  );
}
