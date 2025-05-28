import * as Dialog from '@radix-ui/react-dialog';
import { FieldArray, Form, Formik } from 'formik';
import { XIcon } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { v4 as uuid } from 'uuid';
import * as Yup from 'yup';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import SegmentsPicker from '~/components/forms/SegmentsPicker';

import { DistributionType, IDistribution } from '~/types/Distribution';
import { IRule } from '~/types/Rule';
import {
  FilterableSegment,
  ISegment,
  SegmentOperatorType,
  segmentOperators
} from '~/types/Segment';
import { FilterableVariant, IVariant } from '~/types/Variant';

import { useError } from '~/data/hooks/error';
import { keyValidation } from '~/data/validations';

import MultiDistributionFormInputs from './MultiDistributionForm';
import SingleDistributionFormInput from './SingleDistributionForm';

export const distTypes = [
  {
    id: DistributionType.None,
    name: 'Default Variant / None',
    description: 'The default variant will be returned if any'
  },
  {
    id: DistributionType.Single,
    name: 'Single Variant',
    description: 'Always returns the same variant'
  },
  {
    id: DistributionType.Multi,
    name: 'Multi-Variate',
    description: 'Returns different variants based on percentages'
  }
];

const validationSchema = Yup.object({
  segments: Yup.array()
    .of(
      Yup.object().shape({
        key: keyValidation
      })
    )
    .required()
    .min(1)
});

const computePercentages = (n: number): number[] => {
  const sum = 100 * 100;

  const d = Math.floor(sum / n);
  const remainder = sum - d * n;

  const result = [];
  let i = 0;

  while (++i && i <= n) {
    result.push((i <= remainder ? d + 1 : d) / 100);
  }

  return result;
};

const validRollout = (distributions: IDistribution[]): boolean => {
  const sum = distributions.reduce(function (acc, d) {
    return acc + Number(d.rollout);
  }, 0);

  return sum <= 100;
};

type RuleFormProps = {
  setOpen: (open: boolean) => void;
  onSuccess: () => void;
  createRule: (rule: IRule) => void;
  variants: IVariant[];
  segments: ISegment[];
};

interface Segment {
  segments: FilterableSegment[];
  operator: SegmentOperatorType;
}

export default function RuleForm(props: RuleFormProps) {
  const { setOpen, onSuccess, segments, createRule, variants } = props;

  const { setError, clearError } = useError();

  const [distributionsValid, setDistributionsValid] = useState<boolean>(true);

  const [ruleType, setRuleType] = useState(DistributionType.None);

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(null);

  const [distributions, setDistributions] = useState<IDistribution[]>(() => {
    const percentages = computePercentages(variants.length || 0);

    return (
      variants.map((variant, i) => ({
        variant: variant.key,
        rollout: percentages[i]
      })) || []
    );
  });

  const handleVariantChange = useCallback(
    (variant: FilterableVariant | null) => {
      setSelectedVariant(variant);

      if (variant && ruleType === DistributionType.Single) {
        setDistributions([
          {
            variant: variant.key,
            rollout: 100
          }
        ]);
      } else if (!variant && ruleType === DistributionType.Single) {
        setDistributions([]);
      }
    },
    [ruleType]
  );

  useEffect(() => {
    if (ruleType !== DistributionType.Single) {
      setSelectedVariant(null);
    }
  }, [ruleType]);

  useEffect(() => {
    if (
      ruleType === DistributionType.Multi &&
      distributions.length > 0 &&
      !validRollout(distributions)
    ) {
      setDistributionsValid(false);
    } else {
      setDistributionsValid(true);
    }
  }, [distributions, ruleType]);

  const initialSegmentKeys: FilterableSegment[] = [];

  const handleSubmit = async (values: Segment) => {
    if (values.segments.length === 0) {
      throw new Error('No segments selected');
    }

    const dist = [];
    if (ruleType === DistributionType.Multi && distributions.length > 0) {
      dist.push(
        ...distributions.map((d) => {
          return {
            variant: d.variant,
            rollout: d.rollout
          };
        })
      );
    } else if (ruleType === DistributionType.Single && selectedVariant) {
      dist.push({
        variant: selectedVariant.key,
        rollout: 100
      });
    }

    const segmentKeys = values.segments.map((s) =>
      typeof s === 'string' ? s : s.key
    );

    return createRule({
      segments: segmentKeys,
      segmentOperator: values.operator,
      distributions: dist.map((d) => {
        return {
          ...d,
          id: uuid()
        };
      })
    });
  };

  return (
    <Formik
      initialValues={{
        segments: initialSegmentKeys,
        operator: SegmentOperatorType.OR
      }}
      validationSchema={validationSchema}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            onSuccess();
            clearError();
            setOpen(false);
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
          <Form className="flex h-full flex-col overflow-y-scroll bg-background dark:bg-gray-900 shadow-xl">
            <div className="flex-1">
              <div className="bg-gray-50 dark:bg-gray-800 px-4 py-6 sm:px-6">
                <div className="flex items-start justify-between space-x-3">
                  <div className="space-y-1">
                    <Dialog.Title className="text-lg font-medium text-gray-900 dark:text-gray-100">
                      New Rule
                    </Dialog.Title>
                    <MoreInfo href="https://www.flipt.io/docs/concepts#rules">
                      Learn more about rules
                    </MoreInfo>
                  </div>
                  <div className="flex h-7 items-center">
                    <button
                      type="button"
                      className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
                      onClick={() => setOpen(false)}
                    >
                      <span className="sr-only">Close panel</span>
                      <XIcon className="h-6 w-6" aria-hidden="true" />
                    </button>
                  </div>
                </div>
              </div>
              <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 dark:sm:divide-gray-700 sm:py-0">
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <div>
                    <label
                      htmlFor="segments"
                      className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <FieldArray
                      name="segments"
                      render={(arrayHelpers) => {
                        const handleSegmentAdd = (
                          segment: FilterableSegment
                        ) => {
                          arrayHelpers.push(segment);
                        };

                        const handleSegmentRemove = (index: number) => {
                          arrayHelpers.remove(index);
                        };

                        const handleSegmentReplace = (
                          index: number,
                          segment: FilterableSegment
                        ) => {
                          arrayHelpers.replace(index, segment);
                        };

                        return (
                          <SegmentsPicker
                            segments={segments}
                            segmentAdd={handleSegmentAdd}
                            segmentRemove={handleSegmentRemove}
                            segmentReplace={handleSegmentReplace}
                            selectedSegments={formik.values.segments}
                          />
                        );
                      }}
                    />
                  </div>
                  {formik.values.segments.length > 1 && (
                    <>
                      <div>
                        <label
                          htmlFor="operator"
                          className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                        >
                          Operator
                        </label>
                      </div>
                      <div>
                        <div className="sm:col-span-2">
                          <div className="w-48 space-y-4">
                            {formik.values.segments.length > 1 &&
                              segmentOperators.map((segmentOperator, index) => (
                                <div
                                  className="flex space-x-4 cursor-pointer"
                                  key={index}
                                  onClick={() => {
                                    formik.setFieldValue(
                                      'operator',
                                      segmentOperator.id
                                    );
                                  }}
                                >
                                  <div>
                                    <input
                                      id={segmentOperator.id}
                                      name="operator"
                                      type="radio"
                                      className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400 cursor-pointer"
                                      checked={
                                        segmentOperator.id ===
                                        formik.values.operator
                                      }
                                      value={segmentOperator.id}
                                      readOnly
                                    />
                                  </div>
                                  <div className="mt-1">
                                    <label
                                      htmlFor={segmentOperator.id}
                                      className="block text-sm text-gray-700 dark:text-gray-300 cursor-pointer"
                                    >
                                      {segmentOperator.name}{' '}
                                      <span className="font-light dark:text-gray-400">
                                        {segmentOperator.meta}
                                      </span>
                                    </label>
                                  </div>
                                </div>
                              ))}
                          </div>
                        </div>
                      </div>
                    </>
                  )}
                </div>
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
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
                        {distTypes.map((dist) => (
                          <div
                            key={dist.id}
                            className="relative flex items-start cursor-pointer"
                            onClick={() => {
                              setRuleType(dist.id);
                            }}
                          >
                            <div className="flex h-5 items-center">
                              <input
                                id={dist.id}
                                aria-describedby={`${dist.id}-description`}
                                name="ruleType"
                                type="radio"
                                className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400 cursor-pointer"
                                checked={dist.id === ruleType}
                                value={dist.id}
                                readOnly
                                disabled={
                                  dist.id !== DistributionType.None &&
                                  (!variants || variants.length === 0)
                                }
                              />
                            </div>
                            <div className="ml-3 text-sm">
                              <label
                                htmlFor={dist.id}
                                className="font-medium text-gray-700 dark:text-gray-300 cursor-pointer"
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
                {variants &&
                  variants.length > 0 &&
                  ruleType === DistributionType.Single && (
                    <SingleDistributionFormInput
                      variants={variants}
                      selectedVariant={selectedVariant}
                      setSelectedVariant={handleVariantChange}
                      id="variant"
                    />
                  )}
                {variants &&
                  variants.length > 0 &&
                  ruleType === DistributionType.Multi && (
                    <MultiDistributionFormInputs
                      distributions={distributions}
                      setDistributions={setDistributions}
                    />
                  )}
                {!distributionsValid && ruleType === DistributionType.Multi && (
                  <p className="mt-1 px-4 text-center text-sm text-destructive sm:px-6 sm:py-5">
                    Multi-variate rules must have distributions that add up to
                    100% or less.
                  </p>
                )}
                {(!variants || variants?.length == 0) && (
                  <p className="mt-1 px-4 text-center text-sm text-muted-foreground sm:px-6 sm:py-5">
                    Flag has no variants.
                  </p>
                )}
              </div>
            </div>
            <div className="shrink-0 border-t border-gray-200 dark:border-gray-700 px-4 py-5 sm:px-6">
              <div className="flex justify-end space-x-3">
                <Button variant="secondary" onClick={() => setOpen(false)}>
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  type="submit"
                  className="min-w-[80px]"
                  disabled={
                    !(
                      formik.dirty &&
                      formik.isValid &&
                      distributionsValid &&
                      !formik.isSubmitting
                    )
                  }
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : 'Add'}
                </Button>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
