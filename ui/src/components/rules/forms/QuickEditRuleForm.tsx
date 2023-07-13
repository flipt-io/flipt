import { Field, FieldArray, Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import TextButton from '~/components/forms/buttons/TextButton';
import Combobox from '~/components/forms/Combobox';
import Loading from '~/components/Loading';
import { updateDistribution, updateRule } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IEvaluatable, IVariantRollout } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { FilterableSegment, ISegment } from '~/types/Segment';
import { FilterableVariant } from '~/types/Variant';
import { truncateKey } from '~/utils/helpers';
import { distTypeMulti, distTypes, distTypeSingle } from './RuleForm';

type QuickEditRuleFormProps = {
  onSuccess: () => void;
  flag: IFlag;
  rule: IEvaluatable;
  segments: ISegment[];
};

export const validRollout = (rollouts: IVariantRollout[]): boolean => {
  const sum = rollouts.reduce(function (acc, d) {
    return acc + Number(d.distribution.rollout);
  }, 0);

  return sum <= 100;
};

interface RuleFormValues {
  segmentKey: string;
  rollouts: IVariantRollout[];
}

export default function QuickEditRuleForm(props: QuickEditRuleFormProps) {
  const { onSuccess, flag, rule, segments } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);

  const ruleType = rule.rollouts.length > 1 ? distTypeMulti : distTypeSingle;

  const [selectedSegment, setSelectedSegment] =
    useState<FilterableSegment | null>(() => {
      let selected = segments.find((s) => s.key === rule.segment.key) || null;
      if (selected) {
        return {
          ...selected,
          displayValue: selected.name,
          filterValue: selected.key
        };
      }
      return null;
    });

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (ruleType !== distTypeSingle) return null;

      let selected = rule.rollouts[0].variant;
      if (selected) {
        return {
          ...selected,
          displayValue: selected.name,
          filterValue: selected.id
        };
      }
      return null;
    });

  const handleSubmit = async (values: RuleFormValues) => {
    if (rule.segment.key !== values.segmentKey) {
      // update segment if changed
      try {
        await updateRule(namespace.key, flag.key, rule.id, {
          rank: rule.rank,
          segmentKey: values.segmentKey
        });
      } catch (err) {
        setError(err as Error);
        return;
      }
    }

    if (ruleType === distTypeMulti) {
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
            return updateDistribution(
              namespace.key,
              flag.key,
              rule.id,
              rollout.distribution.id,
              {
                variantId: rollout.distribution.variantId,
                rollout: rollout.distribution.rollout
              }
            );
          }
          return Promise.resolve();
        })
      );
    }
    // TODO: enable once we allow user to change variant of existing single dist rule

    // } else if (ruleType === distTypeSingle && selectedVariant) {
    //   // update variant if changed
    //   if (rule.rollouts[0].distribution.variantId !== selectedVariant.id) {
    //     try {
    //       await updateDistribution(
    //         namespace.key,
    //         flag.key,
    //         rule.id,
    //         rule.rollouts[0].distribution.id,
    //         {
    //           variantId: selectedVariant.id,
    //           rollout: 100
    //         }
    //       );
    //     } catch (err) {
    //       setError(err as Error);
    //       return;
    //     }
    //   }
    // }
  };

  return (
    <Formik
      initialValues={{
        segmentKey: rule.segment.key,
        // variantId: rule.rollouts[0].distribution.variantId,
        rollouts: rule.rollouts
      }}
      validate={(values) => {
        const errors: any = {};
        if (!validRollout(values.rollouts)) {
          errors.rollouts = 'Rollouts must add up to 100%';
        }
        return errors;
      }}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          ?.then(() => {
            clearError();
            setSuccess('Successfully updated rule');
            onSuccess();
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
          <Form className="flex h-full w-full flex-col overflow-y-scroll bg-white">
            <div className="w-full flex-1">
              <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                  <div>
                    <label
                      htmlFor="segmentKey"
                      className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                    >
                      Segment
                    </label>
                  </div>
                  <div className="sm:col-span-2">
                    <Combobox<FilterableSegment>
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
                            <div
                              key={dist.id}
                              className="relative flex items-start"
                            >
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

                {ruleType === distTypeSingle && (
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-2">
                    <div>
                      <label
                        htmlFor="variantKey"
                        className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                      >
                        Variant
                      </label>
                    </div>
                    <div className="sm:col-span-2">
                      <Combobox<FilterableVariant>
                        id="variantKey"
                        name="variantKey"
                        values={flag.variants?.map((v) => ({
                          ...v,
                          filterValue: truncateKey(v.key),
                          displayValue: v.key
                        }))}
                        selected={selectedVariant}
                        setSelected={setSelectedVariant}
                        disabled
                      />
                    </div>
                  </div>
                )}

                {ruleType === distTypeMulti && (
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
                          {formik.values.rollouts &&
                            formik.values.rollouts.length > 0 &&
                            formik.values.rollouts?.map((dist, index) => (
                              <>
                                {dist && (
                                  <div
                                    className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-1"
                                    key={index}
                                  >
                                    <div>
                                      <label
                                        htmlFor={`rollouts.[${index}].distribution.rollout`}
                                        className="block truncate text-right text-sm text-gray-600 sm:mt-px sm:pr-2 sm:pt-2"
                                      >
                                        {dist.variant.key}
                                      </label>
                                    </div>
                                    <div className="relative sm:col-span-1">
                                      <Field
                                        key={index}
                                        type="number"
                                        className="block w-full rounded-md pl-7 pr-12 shadow-sm border-gray-300 focus:ring-violet-300 focus:border-violet-300 sm:text-sm"
                                        value={dist.distribution.rollout}
                                        name={`rollouts.[${index}].distribution.rollout`}
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
                              </>
                            ))}
                        </>
                      )}
                    />
                    {formik.touched.rollouts && formik.errors.rollouts && (
                      <p className="mt-1 px-4 text-center text-sm text-gray-500 sm:px-6 sm:py-5">
                        Multi-variant rules must have distributions that add up
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
