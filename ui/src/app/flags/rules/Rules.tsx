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
import { useFormikContext } from 'formik';
import { SplitSquareVerticalIcon, StarIcon } from 'lucide-react';
import { useCallback, useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { useListSegmentsQuery } from '~/app/segments/segmentsApi';

import { Button, ButtonWithPlus } from '~/components/Button';
import Dropdown from '~/components/Dropdown';
import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import Rule from '~/components/rules/Rule';
import RuleForm from '~/components/rules/RuleForm';
import SingleDistributionFormInput from '~/components/rules/SingleDistributionForm';
import SortableRule from '~/components/rules/SortableRule';

import { IFlag } from '~/types/Flag';
import { IRule } from '~/types/Rule';
import { FilterableVariant, IVariant } from '~/types/Variant';

import { useError } from '~/data/hooks/error';

type RulesProps = {
  flag: IFlag;
  rules: IRule[];
  variants: IVariant[];
};

export function DefaultVariant(props: RulesProps) {
  const { variants, flag } = props;

  const { updateFlag } = useContext(FlagFormContext);

  const [selectedVariant, setSelectedVariant] =
    useState<FilterableVariant | null>(() => {
      if (!flag.defaultVariant) return null;

      const variant = variants?.find((v) => v.key === flag.defaultVariant);
      if (!variant) return null;

      return {
        ...variant,
        displayValue: variant.name || variant.key
      };
    });

  const handleVariantChange = useCallback(
    (variant: FilterableVariant | null) => {
      setSelectedVariant(variant);

      updateFlag({ defaultVariant: variant?.key || null });
    },
    [updateFlag]
  );

  const handleRemove = async () => {
    updateFlag({ defaultVariant: null });
    setSelectedVariant(null);
    return Promise.resolve();
  };

  const formik = useFormikContext<IFlag>();

  return (
    <div className="flex flex-col p-2">
      <div className="w-full items-center space-y-2 rounded-md border bg-background hover:shadow-md hover:shadow-accent sm:flex sm:flex-col px-4 lg:py-2">
        <div className="w-full rounded-t-lg border-b p-2">
          <div className="flex w-full flex-wrap items-center justify-between sm:flex-nowrap">
            <StarIcon className="hidden h-4 w-4 justify-start text-muted-foreground sm:flex" />
            <h3 className="text-sm font-medium text-secondary-foreground">
              Default Rule
            </h3>
            <Dropdown
              size="icon"
              label=""
              kind="dots"
              actions={[
                {
                  id: 'rule-delete',
                  label: 'Delete',
                  variant: 'destructive',
                  disabled: formik.isSubmitting || !selectedVariant,
                  onClick: () => {
                    handleRemove();
                  }
                }
              ]}
            />
          </div>
        </div>

        <div className="flex grow flex-col items-center justify-center sm:ml-2">
          <p className="text-center text-sm text-muted-foreground/75">
            This is the default value that will be returned if no other rules
            match.
          </p>
        </div>
        <div className="flex w-full flex-1 items-center p-2 text-xs lg:p-0">
          <div className="flex grow flex-col items-center justify-center sm:ml-2 md:flex-row md:justify-between">
            <div className="flex w-full flex-col">
              <div className="w-full flex-1">
                <div className="space-y-6 py-6 sm:space-y-0 sm:py-0">
                  {variants && variants.length > 0 && (
                    <SingleDistributionFormInput
                      id="defaultVariant"
                      variants={variants}
                      selectedVariant={selectedVariant}
                      setSelectedVariant={handleVariantChange}
                    />
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function Rules({ flag, variants, rules }: RulesProps) {
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
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });
  const segments = useMemo(
    () => segmentsList.data?.segments || [],
    [segmentsList]
  );

  const showDefaultVariant = variants.length > 0;

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
              <span className="font-medium text-brand">
                {' '}
                position {rules.findIndex((r) => r.id === deletingRule?.id) + 1}
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
          variants={variants}
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
          <div className="sm:flex-auto text-muted-foreground text-sm">
            <p className="mt-1">
              Rules are evaluated in order from top to bottom.
            </p>
            <p className="mt-1">
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
              <div className="w-full border p-4 bg-sidebar">
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
                          rules.map((rule, index) => (
                            <SortableRule
                              key={rule.id}
                              rule={rule}
                              segments={segments}
                              index={index}
                              variants={variants}
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
                          rule={activeRule}
                          segments={segments}
                          variants={variants}
                          index={rules.findIndex((r) => r.id === activeRule.id)}
                        />
                      ) : null}
                    </DragOverlay>
                  </DndContext>
                )}
                {showDefaultVariant && (
                  <DefaultVariant
                    flag={flag}
                    rules={rules}
                    variants={variants}
                  />
                )}
              </div>
            </div>
          ) : (
            <Well>
              <SplitSquareVerticalIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-4">
                No Rules Yet
              </h3>
              <Button
                variant="primary"
                disabled={(flag.variants?.length || 0) < 1}
                aria-label="New Rule"
                onClick={() => {
                  setShowRuleForm(true);
                }}
              >
                Create Rule
              </Button>
            </Well>
          )}
        </div>
      </div>
    </>
  );
}
