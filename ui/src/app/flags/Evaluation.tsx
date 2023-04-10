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
import { InformationCircleIcon } from '@heroicons/react/20/solid';
import { useCallback, useEffect, useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import Button from '~/components/forms/Button';
import Modal from '~/components/Modal';
import EditRuleForm from '~/components/rules/EditRuleForm';
import Rule from '~/components/rules/Rule';
import RuleForm from '~/components/rules/RuleForm';
import SortableRule from '~/components/rules/SortableRule';
import Slideover from '~/components/Slideover';
import { deleteRule, listRules, listSegments, orderRules } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { IDistribution } from '~/types/Distribution';
import { IEvaluatable } from '~/types/Evaluatable';
import { IRule, IRuleList } from '~/types/Rule';
import { ISegment, ISegmentList } from '~/types/Segment';
import { IVariant } from '~/types/Variant';
import { classNames } from '~/utils/helpers';
import { FlagProps } from './FlagProps';

export default function Evaluation() {
  const { flag } = useOutletContext<FlagProps>();

  const [segments, setSegments] = useState<ISegment[]>([]);
  const [rules, setRules] = useState<IEvaluatable[]>([]);

  const [activeRule, setActiveRule] = useState<IEvaluatable | null>(null);

  const [rulesVersion, setRulesVersion] = useState(0);
  const [showRuleForm, setShowRuleForm] = useState<boolean>(false);

  const [showEditRuleForm, setShowEditRuleForm] = useState<boolean>(false);
  const [editingRule, setEditingRule] = useState<IEvaluatable | null>(null);

  const [showDeleteRuleModal, setShowDeleteRuleModal] =
    useState<boolean>(false);
  const [deletingRule, setDeletingRule] = useState<IEvaluatable | null>(null);

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const loadData = useCallback(async () => {
    const segmentList = (await listSegments(
      currentNamespace.key
    )) as ISegmentList;
    const { segments } = segmentList;
    setSegments(segments);

    const ruleList = (await listRules(
      currentNamespace.key,
      flag.key
    )) as IRuleList;

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

      const segment = segments.find(
        (segment: ISegment) => segment.key === rule.segmentKey
      );
      if (!segment) {
        return [];
      }

      return {
        id: rule.id,
        flag,
        segment,
        rank: rule.rank,
        rollouts,
        createdAt: rule.createdAt,
        updatedAt: rule.updatedAt
      };
    });

    setRules(rules);
  }, [currentNamespace.key, flag]);

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
      currentNamespace.key,
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
            deleteRule(currentNamespace.key, flag.key, deletingRule?.id ?? '')
          }
          onSuccess={() => {
            incrementRulesVersion();
          }}
        />
      </Modal>

      {/* rule create form */}
      <Slideover open={showRuleForm} setOpen={setShowRuleForm}>
        <RuleForm
          flag={flag}
          rank={(rules?.length || 0) + 1}
          segments={segments}
          setOpen={setShowRuleForm}
          rulesChanged={incrementRulesVersion}
        />
      </Slideover>

      {/* rule edit form */}
      {editingRule && (
        <Slideover open={showEditRuleForm} setOpen={setShowEditRuleForm}>
          <EditRuleForm
            rule={editingRule}
            setOpen={setShowEditRuleForm}
            onSuccess={() => {
              incrementRulesVersion();
              setShowEditRuleForm(false);
            }}
          />
        </Slideover>
      )}

      {/* rules */}
      <div className="my-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-lg font-medium leading-6 text-gray-900">
              Rules
            </h1>
            <p className="mt-1 text-sm text-gray-600">
              Enable rich targeting and segmentation for evaluating your flags
            </p>
          </div>
          {rules && rules.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                primary
                type="button"
                onClick={() => setShowRuleForm(true)}
              >
                New Rule
              </Button>
            </div>
          )}
        </div>
        <div className="mt-10">
          {rules && rules.length > 0 ? (
            <div className="flex lg:space-x-5">
              <div className="hidden w-1/4 flex-col space-y-7 pr-3 lg:flex">
                <p className="text-sm text-gray-500">
                  Rules are evaluated in order from{' '}
                  <span className="font-semibold">top to bottom</span>. The
                  first rule that matches will be applied.
                </p>
                <p className="text-sm text-gray-500">
                  <InformationCircleIcon className="mr-1 inline-block h-4 w-4 text-violet-300" />
                  You can re-arrange rules by{' '}
                  <span className="font-semibold">dragging and dropping</span>{' '}
                  them into place.
                </p>
              </div>
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
                    className={classNames(
                      'w-full space-y-5 p-5 lg:w-3/4',
                      activeRule ? 'bg-violet-50' : 'bg-gray-50'
                    )}
                  >
                    {rules &&
                      rules.length > 0 &&
                      rules.map((rule) => (
                        <SortableRule
                          key={rule.id}
                          namespace={currentNamespace}
                          rule={rule}
                          onEdit={() => {
                            setEditingRule(rule);
                            setShowEditRuleForm(true);
                          }}
                          onDelete={() => {
                            setDeletingRule(rule);
                            setShowDeleteRuleModal(true);
                          }}
                        />
                      ))}
                  </ul>
                </SortableContext>
                <DragOverlay>
                  {activeRule ? (
                    <Rule namespace={currentNamespace} rule={activeRule} />
                  ) : null}
                </DragOverlay>
              </DndContext>
            </div>
          ) : (
            <EmptyState
              text="New Rule"
              onClick={() => {
                setEditingRule(null);
                setShowRuleForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
