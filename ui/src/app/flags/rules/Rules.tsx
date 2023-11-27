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
import { PlusIcon } from '@heroicons/react/24/outline';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useOutletContext } from 'react-router-dom';
import { FlagProps } from '~/app/flags/FlagProps';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/buttons/Button';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import RuleForm from '~/components/rules/forms/RuleForm';
import Rule from '~/components/rules/Rule';
import SortableRule from '~/components/rules/SortableRule';
import Slideover from '~/components/Slideover';
import { deleteRule, listRules, orderRules } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { IDistribution } from '~/types/Distribution';
import { IEvaluatable } from '~/types/Evaluatable';
import { FlagType } from '~/types/Flag';
import { IRule, IRuleList } from '~/types/Rule';
import { ISegment, SegmentOperatorType } from '~/types/Segment';
import { IVariant } from '~/types/Variant';

export default function Rules() {
  const { flag } = useOutletContext<FlagProps>();

  const [rules, setRules] = useState<IEvaluatable[]>([]);

  const [activeRule, setActiveRule] = useState<IEvaluatable | null>(null);

  const [rulesVersion, setRulesVersion] = useState(0);
  const [showRuleForm, setShowRuleForm] = useState<boolean>(false);

  const [showDeleteRuleModal, setShowDeleteRuleModal] =
    useState<boolean>(false);
  const [deletingRule, setDeletingRule] = useState<IEvaluatable | null>(null);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const navigate = useNavigate();

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);
  const listSegments = useListSegmentsQuery(namespace.key);
  const segments = useMemo(
    () => listSegments.data?.segments || [],
    [listSegments]
  );

  const loadData = useCallback(async () => {
    // TODO: move to redux
    const ruleList = (await listRules(namespace.key, flag.key)) as IRuleList;

    const rules = ruleList.rules.flatMap((rule: IRule) => {
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
      // TODO(yquansah): Should be removed once there are no more references to `segmentKey`.
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

      // If there are no ruleSegments return an empty array.
      if (ruleSegments.length === 0) {
        return [];
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

    setRules(rules);
  }, [namespace.key, flag, segments]);

  const incrementRulesVersion = () => {
    setRulesVersion(rulesVersion + 1);
  };

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  const reorderRules = (rules: IEvaluatable[]) => {
    orderRules(
      namespace.key,
      flag.key,
      rules.map((rule) => rule.id)
    )
      .then(() => {
        incrementRulesVersion();
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
      setRules(reordered);
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

  useEffect(() => {
    loadData();
  }, [loadData, rulesVersion]);

  useEffect(() => {
    if (flag.type === FlagType.BOOLEAN) {
      setError('Boolean flags do not support evaluation rules');
      navigate(`/namespaces/${namespace.key}/flags/${flag.key}`);
    }
  }, [flag.type, navigate, setError, namespace.key, flag.key]);

  return (
    <>
      {/* rule delete modal */}
      <Modal open={showDeleteRuleModal} setOpen={setShowDeleteRuleModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete this rule at
              <span className="text-violet-500 font-medium">
                {' '}
                position {deletingRule?.rank}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Rule"
          setOpen={setShowDeleteRuleModal}
          handleDelete={() =>
            deleteRule(namespace.key, flag.key, deletingRule?.id ?? '')
          }
          onSuccess={incrementRulesVersion}
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
            incrementRulesVersion();
            setShowRuleForm(false);
          }}
        />
      </Slideover>

      {/* rules */}
      <div className="mt-2">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="text-gray-500 mt-1 text-sm">
              Enable rich targeting and segmentation for evaluating your flags
            </p>
          </div>
          {rules && rules.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                primary
                type="button"
                onClick={() => setShowRuleForm(true)}
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
              >
                <PlusIcon
                  className="text-white -ml-1.5 mr-1 h-5 w-5"
                  aria-hidden="true"
                />
                New Rule
              </Button>
            </div>
          )}
        </div>
        <div className="mt-10">
          {rules && rules.length > 0 ? (
            <div className="flex lg:space-x-5">
              <div className="hidden w-1/4 flex-col space-y-7 pr-3 lg:flex">
                <p className="text-gray-700 text-sm font-light">
                  Rules are evaluated in order from{' '}
                  <span className="font-semibold">top to bottom</span>. The
                  first rule that matches will be applied.
                </p>
                <p className="text-gray-700 text-sm font-light">
                  Rules can be rearranged by clicking on a rule header and{' '}
                  <span className="font-semibold">dragging and dropping</span>{' '}
                  it into place.
                </p>
              </div>
              <div
                className="border-gray-200 pattern-boxes w-full border p-4 pattern-bg-gray-50 pattern-gray-100 pattern-opacity-100 pattern-size-2 dark:pattern-bg-black dark:pattern-gray-900
  lg:w-3/4 lg:p-6"
              >
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
                    <ul role="list" className="flex-col space-y-5 md:flex">
                      {rules &&
                        rules.length > 0 &&
                        rules.map((rule) => (
                          <SortableRule
                            key={rule.id}
                            flag={flag}
                            rule={rule}
                            segments={segments}
                            onSuccess={incrementRulesVersion}
                            onDelete={() => {
                              setDeletingRule(rule);
                              setShowDeleteRuleModal(true);
                            }}
                            readOnly={readOnly}
                          />
                        ))}
                    </ul>
                  </SortableContext>
                  <DragOverlay>
                    {activeRule ? (
                      <Rule flag={flag} rule={activeRule} segments={segments} />
                    ) : null}
                  </DragOverlay>
                </DndContext>
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
