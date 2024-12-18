import {
  closestCenter,
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { StarIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router-dom';
import {
  useDeleteRuleMutation,
  useListRulesQuery,
  useOrderRulesMutation
} from '~/app/flags/rulesApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import EmptyState from '~/components/EmptyState';
import { ButtonWithPlus, TextButton } from '~/components/Button';

import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import RuleForm from '~/components/rules/forms/RuleForm';
import Rule from '~/components/rules/Rule';
import SortableRule from '~/components/rules/SortableRule';
import Slideover from '~/components/Slideover';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IDistribution } from '~/types/Distribution';
import { IEvaluatable } from '~/types/Evaluatable';
import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { ISegment, SegmentOperatorType } from '~/types/Segment';
import {
  FilterableVariant,
  IVariant,
  toFilterableVariant
} from '~/types/Variant';
import { useUpdateFlagMutation } from '~/app/flags/flagsApi';
import { INamespace } from '~/types/Namespace';
import SingleDistributionFormInput from '~/components/rules/forms/SingleDistributionForm';

type RulesProps = {
  flag: IFlag;
};

export function DefaultVariant(props: RulesProps) {
  const { flag } = props;

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace) as INamespace;
  const readOnly = useSelector(selectReadonly);

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      return toFilterableVariant(flag.defaultVariant);
    });

  const [updateFlag] = useUpdateFlagMutation();

  const handleRemove = async () => {
    try {
      setSelectedVariant(null);
      await updateFlag({
        namespaceKey: namespace.key,
        flagKey: flag.key,
        values: {
          ...flag,
          defaultVariant: undefined
        }
      });
      clearError();
      setSuccess('Successfully removed default variant');
    } catch (err) {
      setError(err);
    }
  };

  const handleSubmit = async () => {
    await updateFlag({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      values: {
        ...flag,
        defaultVariant: {
          id: selectedVariant?.id ?? ''
        } as IVariant
      }
    });
  };

  return (
    <Formik
      initialValues={{
        defaultVariant: selectedVariant
      }}
      onSubmit={(_, { setSubmitting }) => {
        handleSubmit()
          .then(() => {
            clearError();
            setSuccess('Successfully updated default variant');
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
          <div className="flex flex-col p-2">
            <div className="w-full items-center space-y-2 rounded-md border border-violet-300 bg-background shadow-md shadow-violet-100 hover:shadow-violet-200 sm:flex sm:flex-col lg:px-4 lg:py-2">
              <div className="w-full rounded-t-lg border-b border-gray-200 p-2">
                <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
                  <StarIcon className="hidden h-4 w-4 justify-start text-gray-400 hover:text-violet-300 sm:flex" />
                  <h3 className="text-sm font-normal leading-6 text-gray-700">
                    Default Rule
                  </h3>
                  <span className="hidden h-4 w-4 justify-end sm:flex" />
                </div>
              </div>

              <div className="flex grow flex-col items-center justify-center sm:ml-2">
                <p className="text-center text-sm font-light text-gray-600">
                  This is the default value that will be returned if no other
                  rules match.
                </p>
              </div>
              <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
                <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
                  <Form className="flex w-full flex-col">
                    <div className="w-full flex-1">
                      <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                        {flag.variants && flag.variants.length > 0 && (
                          <SingleDistributionFormInput
                            id="variant-default"
                            variants={flag.variants}
                            selectedVariant={selectedVariant}
                            setSelectedVariant={setSelectedVariant}
                          />
                        )}
                      </div>
                    </div>
                    <div className="flex-shrink-0 py-1">
                      <div className="flex justify-end space-x-3">
                        <TextButton
                          className="min-w-[80px]"
                          disabled={
                            formik.isSubmitting ||
                            !flag.defaultVariant ||
                            readOnly
                          }
                          onClick={() => handleRemove()}
                        >
                          Remove
                        </TextButton>
                        <TextButton
                          disabled={
                            formik.isSubmitting ||
                            flag.defaultVariant?.id == selectedVariant?.id ||
                            readOnly
                          }
                          onClick={() => {
                            formik.resetForm();
                            setSelectedVariant(
                              toFilterableVariant(flag.defaultVariant)
                            );
                          }}
                        >
                          Reset
                        </TextButton>
                        <TextButton
                          type="submit"
                          className="min-w-[80px]"
                          disabled={
                            !formik.isValid ||
                            formik.isSubmitting ||
                            flag.defaultVariant?.id == selectedVariant?.id ||
                            readOnly
                          }
                        >
                          {formik.isSubmitting ? (
                            <Loading isPrimary />
                          ) : (
                            'Update'
                          )}
                        </TextButton>
                      </div>
                    </div>
                  </Form>
                </div>
              </div>
            </div>
          </div>
        );
      }}
    </Formik>
  );
}

export default function Rules() {
  const { flag } = useOutletContext<RulesProps>();

  const [activeRule, setActiveRule] = useState<IEvaluatable | null>(null);

  const [showRuleForm, setShowRuleForm] = useState<boolean>(false);

  const [showDeleteRuleModal, setShowDeleteRuleModal] =
    useState<boolean>(false);
  const [deletingRule, setDeletingRule] = useState<IEvaluatable | null>(null);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);
  const segmentsList = useListSegmentsQuery(namespace.key);
  const segments = useMemo(
    () => segmentsList.data?.segments || [],
    [segmentsList]
  );

  const showDefaultVariant = flag.variants && flag.variants.length > 0;

  const [deleteRule] = useDeleteRuleMutation();
  const [orderRules] = useOrderRulesMutation();

  const rulesList = useListRulesQuery({
    namespaceKey: namespace.key,
    flagKey: flag.key
  });

  const ruleList = useMemo(() => rulesList.data?.rules || [], [rulesList]);

  const rules = useMemo(() => {
    return ruleList.flatMap((rule: IRule) => {
      const rollouts = rule.distributions.flatMap(
        (distribution: IDistribution) => {
          const variant = flag?.variants?.find(
            (variant: IVariant) => variant.id === distribution.variantId
          );

          if (!variant) {
            return [];
          }

          return {
            variant,
            distribution
          };
        }
      );

      const ruleSegments: ISegment[] = [];

      const size = rule.segmentKeys ? rule.segmentKeys.length : 0;

      // Combine both segment and segments for legacy purposes.
      for (let i = 0; i < size; i++) {
        const ruleSegment = rule.segmentKeys && rule.segmentKeys[i];
        const segment = segments.find(
          (segment: ISegment) => ruleSegment === segment.key
        );
        if (segment) {
          ruleSegments.push(segment);
        }
      }

      const segment = segments.find(
        (segment: ISegment) => segment.key === rule.segmentKey
      );

      if (segment) {
        ruleSegments.push(segment);
      }

      const operator = rule.segmentOperator
        ? rule.segmentOperator
        : SegmentOperatorType.OR;

      return {
        id: rule.id,
        flag,
        segments: ruleSegments,
        operator,
        rank: rule.rank,
        rollouts,
        createdAt: rule.createdAt,
        updatedAt: rule.updatedAt
      };
    });
  }, [flag, segments, ruleList]);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  const reorderRules = (rules: IEvaluatable[]) => {
    orderRules({
      namespaceKey: namespace.key,
      flagKey: flag.key,
      ruleIds: rules.map((rule) => rule.id)
    })
      .unwrap()
      .then(() => {
        clearError();
        setSuccess('Successfully reordered rules');
      })
      .catch((err) => {
        setError(err);
      });
  };

  // disabling eslint due to this being a third-party event type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragEnd = (event: { active: any; over: any }) => {
    const { active, over } = event;

    if (active.id !== over.id) {
      const reordered = (function (rules: IEvaluatable[]) {
        const oldIndex = rules.findIndex((rule) => rule.id === active.id);
        const newIndex = rules.findIndex((rule) => rule.id === over.id);

        return arrayMove(rules, oldIndex, newIndex);
      })(rules);

      reorderRules(reordered);
    }

    setActiveRule(null);
  };

  // disabling eslint due to this being a third-party event type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragStart = (event: { active: any }) => {
    const { active } = event;
    const rule = rules.find((rule) => rule.id === active.id);
    if (rule) {
      setActiveRule(rule);
    }
  };

  if (segmentsList.isLoading || rulesList.isLoading) {
    return <Loading />;
  }

  return (
    <>
      {/* rule delete modal */}
      <Modal open={showDeleteRuleModal} setOpen={setShowDeleteRuleModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete this rule at
              <span className="font-medium text-violet-500">
                {' '}
                position {deletingRule?.rank}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Rule"
          setOpen={setShowDeleteRuleModal}
          handleDelete={() =>
            deleteRule({
              namespaceKey: namespace.key,
              flagKey: flag.key,
              ruleId: deletingRule?.id ?? ''
            }).unwrap()
          }
        />
      </Modal>

      {/* rule create form */}
      <Slideover open={showRuleForm} setOpen={setShowRuleForm}>
        <RuleForm
          flag={flag}
          rank={(rules?.length || 0) + 1}
          segments={segments}
          setOpen={setShowRuleForm}
          onSuccess={() => {
            setShowRuleForm(false);
          }}
        />
      </Slideover>

      {/* rules */}
      <div className="mt-2">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="mt-1 text-sm text-gray-500">
              Rules are evaluated in order from top to bottom.
            </p>
            <p className="mt-1 text-sm text-gray-500">
              Rules can be rearranged by clicking on the header and dragging and
              dropping it into place.
            </p>
          </div>
          {((rules && rules.length > 0) || showDefaultVariant) && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <ButtonWithPlus
                variant="primary"
                type="button"
                onClick={() => setShowRuleForm(true)}
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
              >
                New Rule
              </ButtonWithPlus>
            </div>
          )}
        </div>
        <div className="mt-10">
          {(rules && rules.length > 0) || showDefaultVariant ? (
            <div className="flex">
              <div className="pattern-boxes w-full border border-gray-200 p-4 pattern-bg-gray-solid50 pattern-gray-solid100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-gray-solid lg:p-6">
                {rules && rules.length > 0 && (
                  <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragStart={onDragStart}
                    onDragEnd={onDragEnd}
                  >
                    <SortableContext
                      items={rules.map((rule) => rule.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      <ul
                        role="list"
                        className="flex-col space-y-6 p-2 md:flex"
                      >
                        {rules &&
                          rules.length > 0 &&
                          rules.map((rule) => (
                            <SortableRule
                              key={rule.id}
                              flag={flag}
                              rule={rule}
                              segments={segments}
                              onDelete={() => {
                                setDeletingRule(rule);
                                setShowDeleteRuleModal(true);
                              }}
                              onSuccess={clearError}
                              readOnly={readOnly}
                            />
                          ))}
                      </ul>
                    </SortableContext>
                    <DragOverlay>
                      {activeRule ? (
                        <Rule
                          flag={flag}
                          rule={activeRule}
                          segments={segments}
                        />
                      ) : null}
                    </DragOverlay>
                  </DndContext>
                )}
                {showDefaultVariant && <DefaultVariant flag={flag} />}
              </div>
            </div>
          ) : (
            <EmptyState
              text="New Rule"
              disabled={readOnly}
              onClick={() => {
                setShowRuleForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
