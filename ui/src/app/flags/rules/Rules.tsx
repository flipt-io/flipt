import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy
} from '@dnd-kit/sortable';
import { StarIcon } from '@heroicons/react/24/outline';
import { useFormikContext } from 'formik';
import { useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { v4 as uuid } from 'uuid';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { ButtonWithPlus, TextButton } from '~/components/Button';
import EmptyState from '~/components/EmptyState';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import Rule from '~/components/rules/Rule';
import RuleForm from '~/components/rules/RuleForm';
import SingleDistributionFormInput from '~/components/rules/SingleDistributionForm';
import SortableRule from '~/components/rules/SortableRule';

import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { FilterableVariant, toFilterableVariant } from '~/types/Variant';

import { useError } from '~/data/hooks/error';

type RulesProps = {
  flag: IFlag;
  rules: IRule[];
};

export function DefaultVariant(props: RulesProps) {
  const { flag } = props;

  const { updateFlag } = useContext(FlagFormContext);

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      const variant = toFilterableVariant(
        flag.variants?.find((variant) => {
          return variant.key == flag.defaultVariant;
        }) || null
      );
      return variant || null;
    });

  const handleRemove = async () => {
    updateFlag({ defaultVariant: null });
    setSelectedVariant(null);
    return Promise.resolve();
  };

  const formik = useFormikContext<IFlag>();

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
            This is the default value that will be returned if no other rules
            match.
          </p>
        </div>
        <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
          <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
            <div className="flex w-full flex-col">
              <div className="w-full flex-1">
                <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                  {flag.variants && flag.variants.length > 0 && (
                    <SingleDistributionFormInput
                      id="defaultVariant"
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
                    disabled={formik.isSubmitting || !flag.defaultVariant}
                    onClick={() => handleRemove()}
                  >
                    Remove
                  </TextButton>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function Rules({ flag, rules }: RulesProps) {
  const [activeRule, setActiveRule] = useState<IRule | null>(null);
  const [deletingRule, setDeletingRule] = useState<IRule | null>(null);

  const [showRuleForm, setShowRuleForm] = useState<boolean>(false);

  const [showDeleteRuleModal, setShowDeleteRuleModal] =
    useState<boolean>(false);

  const { clearError } = useError();

  const ruleFormRef = useRef(null);

  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);

  const segmentsList = useListSegmentsQuery({
    environmentKey: environment.name,
    namespaceKey: namespace.key
  });
  const segments = useMemo(
    () => segmentsList.data?.segments || [],
    [segmentsList]
  );

  const showDefaultVariant = flag.variants && flag.variants.length > 0;

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  const { setRules, deleteRule, createRule } = useContext(FlagFormContext);

  // disabling eslint due to this being a third-party event type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragEnd = (event: { active: any; over: any }) => {
    const { active, over } = event;

    if (active.id !== over.id) {
      const reordered = (function (rules: IRule[]) {
        const oldIndex = rules.findIndex((rule) => rule.id === active.id);
        const newIndex = rules.findIndex((rule) => rule.id === over.id);

        return arrayMove(rules, oldIndex, newIndex);
      })(rules);

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

  if (segmentsList.isLoading) {
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
          handleDelete={() => {
            if (!deletingRule) {
              return Promise.resolve();
            }
            deleteRule(deletingRule);
            return Promise.resolve();
          }}
        />
      </Modal>

      {/* rule create form */}
      <Slideover
        open={showRuleForm}
        setOpen={setShowRuleForm}
        ref={ruleFormRef}
      >
        <RuleForm
          flag={flag}
          rank={(rules?.length || 0) + 1}
          segments={segments}
          setOpen={setShowRuleForm}
          createRule={createRule}
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
                      items={rules.map((rule) => rule.id!)}
                      strategy={verticalListSortingStrategy}
                    >
                      <ul
                        role="list"
                        className="flex-col space-y-6 p-2 md:flex"
                      >
                        {rules &&
                          rules.map((rule) => (
                            <SortableRule
                              key={rule.id}
                              flag={flag}
                              rule={rule}
                              segments={segments}
                              onDelete={() => {
                                setActiveRule(null);
                                setDeletingRule(rule);
                                setShowDeleteRuleModal(true);
                              }}
                              onSuccess={clearError}
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
                {showDefaultVariant && (
                  <DefaultVariant flag={flag} rules={rules} />
                )}
              </div>
            </div>
          ) : (
            <EmptyState
              text="New Rule"
              onClick={(e) => {
                e.preventDefault();
                setShowRuleForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
