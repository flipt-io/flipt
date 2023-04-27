import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import * as Yup from 'yup';
import Button from '~/components/forms/Button';
import Combobox, { ISelectable } from '~/components/forms/Combobox';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { createDistribution, createRule } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation } from '~/data/validations';
import { IDistributionVariant } from '~/types/Distribution';
import { IFlag } from '~/types/Flag';
import { ISegment } from '~/types/Segment';
import { SelectableVariant } from '~/types/Variant';
import { truncateKey } from '~/utils/helpers';
import MultiDistributionFormInputs from './distributions/MultiDistributionForm';
import SingleDistributionFormInput from './distributions/SingleDistributionForm';

type RuleFormProps = {
  setOpen: (open: boolean) => void;
  rulesChanged: () => void;
  flag: IFlag;
  rank: number;
  segments: ISegment[];
};

const distTypes = [
  {
    id: 'single',
    name: 'Single Variant',
    description: 'Always returns the same variant'
  },
  {
    id: 'multi',
    name: 'Multi-Variant',
    description: 'Returns different variants based on percentages'
  }
];

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

const validRollout = (distributions: IDistributionVariant[]): boolean => {
  const sum = distributions.reduce(function (acc, d) {
    return acc + Number(d.rollout);
  }, 0);

  return sum <= 100;
};

type SelectableSegment = ISegment & ISelectable;

export default function RuleForm(props: RuleFormProps) {
  const { setOpen, rulesChanged, flag, rank, segments } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const [distributionsValid, setDistributionsValid] = useState<boolean>(true);

  const [ruleType, setRuleType] = useState('single');

  const [selectedSegment, setSelectedSegment] =
    useState<SelectableSegment | null>(null);
  const [selectedVariant, setSelectedVariant] =
    useState<SelectableVariant | null>(null);

  const [distributions, setDistributions] = useState(() => {
    const percentages = computePercentages(flag.variants?.length || 0);

    return flag.variants?.map((variant, i) => ({
      variantId: variant.id,
      variantKey: variant.key,
      rollout: percentages[i]
    }));
  });

  useEffect(() => {
    if (ruleType === 'multi' && distributions && !validRollout(distributions)) {
      setDistributionsValid(false);
    } else {
      setDistributionsValid(true);
    }
  }, [distributions, ruleType]);

  const handleSubmit = async () => {
    if (!selectedSegment) {
      throw new Error('No segment selected');
    }

    const rule = await createRule(currentNamespace.key, flag.key, {
      flagKey: flag.key,
      segmentKey: selectedSegment.key,
      rank
    });

    if (ruleType === 'multi') {
      const distPromises = distributions?.map((dist: IDistributionVariant) =>
        createDistribution(currentNamespace.key, flag.key, rule.id, {
          variantId: dist.variantId,
          rollout: dist.rollout
        })
      );
      if (distPromises) await Promise.all(distPromises);
    } else {
      if (selectedVariant) {
        // we allow creating rules without variants

        await createDistribution(currentNamespace.key, flag.key, rule.id, {
          variantId: selectedVariant.id,
          rollout: 100
        });
      }
    }
  };

  return (
    <Formik
      initialValues={{
        segmentKey: selectedSegment?.key || ''
      }}
      validationSchema={Yup.object({
        segmentKey: keyValidation
      })}
      onSubmit={(_, { setSubmitting }) => {
        handleSubmit()
          .then(() => {
            rulesChanged();
            clearError();
            setSuccess('Successfully created rule');
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
          <Form className="flex h-full flex-col overflow-y-scroll bg-white shadow-xl">
            <div className="flex-1">
              <div className="bg-gray-50 px-4 py-6 sm:px-6">
                <div className="flex items-start justify-between space-x-3">
                  <div className="space-y-1">
                    <Dialog.Title className="text-lg font-medium text-gray-900">
                      New Rule
                    </Dialog.Title>
                    <MoreInfo href="https://www.flipt.io/docs/concepts#rules">
                      Learn more about rules
                    </MoreInfo>
                  </div>
                  <div className="flex h-7 items-center">
                    <button
                      type="button"
                      className="text-gray-400 hover:text-gray-500"
                      onClick={() => setOpen(false)}
                    >
                      <span className="sr-only">Close panel</span>
                      <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                    </button>
                  </div>
                </div>
              </div>
              <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                  <div>
                    <label
                      htmlFor="segmentKey"
                      className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <Combobox<SelectableSegment>
                      id="segmentKey"
                      name="segmentKey"
                      placeholder="Select or search for a segment"
                      values={segments.map((s) => ({
                        ...s,
                        filterValue: truncateKey(s.key),
                        displayValue: s.name
                      }))}
                      selected={selectedSegment}
                      setSelected={setSelectedSegment}
                    />
                  </div>
                </div>
                {flag.variants && (
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
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
                          {distTypes.map((dist) => (
                            <div
                              key={dist.id}
                              className="relative flex items-start"
                            >
                              <div className="flex h-5 items-center">
                                <input
                                  id={dist.id}
                                  aria-describedby={`${dist.id}-description`}
                                  name="ruleType"
                                  type="radio"
                                  className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                                  onChange={() => {
                                    setRuleType(dist.id);
                                  }}
                                  checked={dist.id === ruleType}
                                  value={dist.id}
                                />
                              </div>
                              <div className="ml-3 text-sm">
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
                )}
                {flag.variants && ruleType === 'single' && (
                  <SingleDistributionFormInput
                    variants={flag.variants}
                    selectedVariant={selectedVariant}
                    setSelectedVariant={setSelectedVariant}
                  />
                )}
                {flag.variants && ruleType === 'multi' && (
                  <MultiDistributionFormInputs
                    distributions={distributions}
                    setDistributions={setDistributions}
                  />
                )}
                {!distributionsValid && (
                  <p className="mt-1 px-4 text-center text-sm text-gray-500 sm:px-6 sm:py-5">
                    Multi-variant rules must have distributions that add up to
                    100% or less.
                  </p>
                )}
                {(!flag.variants || flag.variants?.length == 0) && (
                  <p className="mt-1 px-4 text-center text-sm text-gray-500 sm:px-6 sm:py-5">
                    Flag{' '}
                    <Link to=".." className="text-violet-500">
                      {truncateKey(flag.key)}
                    </Link>{' '}
                    has no variants. You can add variants in the details
                    section.
                  </p>
                )}
              </div>
            </div>
            <div className="flex-shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
              <div className="flex justify-end space-x-3">
                <Button onClick={() => setOpen(false)}>Cancel</Button>
                <Button
                  primary
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
                  {formik.isSubmitting ? <Loading isPrimary /> : 'Create'}
                </Button>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
