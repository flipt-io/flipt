import { SlidersHorizontalIcon, XIcon } from 'lucide-react';
import { useContext, useMemo, useRef, useState } from 'react';
import { useSelector } from 'react-redux';

import { selectViewMode } from '~/app/preferences/preferencesSlice';

import { Button, ButtonWithPlus } from '~/components/Button';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import VariantForm from '~/components/variants/VariantForm';

import { ViewMode } from '~/types/Preferences';
import { IVariant } from '~/types/Variant';

function VariantCard({
  variant,
  onEdit,
  onDelete
}: {
  variant: IVariant;
  onEdit: () => void;
  onDelete: () => void;
}) {
  // Check if variant has a non-empty attachment
  const hasAttachment =
    variant.attachment && Object.keys(variant.attachment).length > 0;

  return (
    <div className="relative flex flex-col rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-950 overflow-hidden shadow-sm hover:shadow-md">
      <div className="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 p-3">
        <div className="flex items-center space-x-2">
          <span className="p-1.5 rounded-md bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-100">
            <SlidersHorizontalIcon className="h-4 w-4" />
          </span>
          <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">
            Variant
          </h3>
        </div>
        <button
          onClick={(e) => {
            e.preventDefault();
            onDelete();
          }}
          className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
        >
          <XIcon className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 p-4 space-y-3" onClick={onEdit}>
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
            Key
          </span>
          <code className="text-sm font-mono text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
            {variant.key}
          </code>
        </div>

        {variant.name && (
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
              Name
            </span>
            <span className="text-sm font-medium text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
              {variant.name}
            </span>
          </div>
        )}

        {hasAttachment && (
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
              Attachment
            </span>
            <span className="text-sm font-medium text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
              <span className="text-gray-500 dark:text-gray-400 text-xs">
                {Object.keys(variant.attachment || {}).length} fields
              </span>
            </span>
          </div>
        )}

        {variant.description && (
          <div className="pt-2 mt-2 border-t border-gray-100 dark:border-gray-800">
            <span className="text-xs font-medium uppercase text-gray-500 dark:text-gray-400 block mb-1">
              Description
            </span>
            <p className="text-sm text-gray-700 dark:text-gray-300 line-clamp-2">
              {variant.description}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

function VariantTable({
  variants,
  onEdit,
  onDelete
}: {
  variants: IVariant[];
  onEdit: (variant: IVariant) => void;
  onDelete: (variant: IVariant) => void;
}) {
  return (
    <div className="mt-4 overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-800">
          <tr>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Key
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Name
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Attachment
            </th>
            <th
              scope="col"
              className="px-3 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Description
            </th>
            <th scope="col" className="relative px-3 py-3">
              <span className="sr-only">Actions</span>
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {variants.map((variant) => {
            const hasAttachment =
              variant.attachment && Object.keys(variant.attachment).length > 0;

            return (
              <tr
                key={variant.key}
                className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
                onClick={() => onEdit(variant)}
              >
                <td className="px-3 py-4 whitespace-nowrap text-sm">
                  <code className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                    {variant.key}
                  </code>
                </td>
                <td className="px-3 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">
                  {variant.name || (
                    <span className="text-gray-400 dark:text-gray-500">—</span>
                  )}
                </td>
                <td className="px-3 py-4 whitespace-nowrap text-sm">
                  {hasAttachment ? (
                    <span className="bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-500 dark:text-gray-400 text-xs">
                      {Object.keys(variant.attachment || {}).length} fields
                    </span>
                  ) : (
                    <span className="text-gray-400 dark:text-gray-500">—</span>
                  )}
                </td>
                <td className="px-3 py-4 text-sm text-gray-700 dark:text-gray-300">
                  {variant.description ? (
                    <p className="line-clamp-1">{variant.description}</p>
                  ) : (
                    <span className="text-gray-400 dark:text-gray-500">—</span>
                  )}
                </td>
                <td className="px-3 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onDelete(variant);
                    }}
                    className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400"
                  >
                    <XIcon className="h-4 w-4" />
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

type VariantsProps = {
  variants: IVariant[];
};

export default function Variants({ variants }: VariantsProps) {
  const [showVariantForm, setShowVariantForm] = useState<boolean>(false);
  const [editingVariant, setEditingVariant] = useState<IVariant | null>(null);
  const [showDeleteVariantModal, setShowDeleteVariantModal] =
    useState<boolean>(false);
  const [deletingVariant, setDeletingVariant] = useState<IVariant | null>(null);

  const variantFormRef = useRef(null);

  const { deleteVariant } = useContext(FlagFormContext);

  // Get user's view mode preference
  const viewMode = useSelector(selectViewMode);

  // Determine if we should use table view based on preference or count
  const useTableView = useMemo(() => {
    if (viewMode === ViewMode.TABLE) return true;
    if (viewMode === ViewMode.CARDS) return false;
    // Default auto behavior - table for more than 6 items
    return variants.length > 6;
  }, [variants.length, viewMode]);

  return (
    <>
      {/* variant edit form */}
      <Slideover
        open={showVariantForm}
        setOpen={setShowVariantForm}
        ref={variantFormRef}
      >
        <VariantForm
          ref={variantFormRef}
          variant={editingVariant}
          setOpen={setShowVariantForm}
          onSuccess={() => {
            setShowVariantForm(false);
          }}
        />
      </Slideover>

      {/* variant delete modal */}
      <Modal open={showDeleteVariantModal} setOpen={setShowDeleteVariantModal}>
        <DeletePanel
          panelMessage={
            <>
              Are you sure you want to delete the variant{' '}
              <span className="font-medium text-violet-500 dark:text-violet-400">
                {deletingVariant?.key}
              </span>
              ? This action cannot be undone.
            </>
          }
          panelType="Variant"
          setOpen={setShowDeleteVariantModal}
          handleDelete={() => {
            try {
              deleteVariant(deletingVariant!);
            } catch (e) {
              return Promise.reject(e);
            }
            setDeletingVariant(null);
            return Promise.resolve();
          }}
        />
      </Modal>

      {/* variants */}
      <div className="mt-2 min-w-full">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-300">
              Return different values based on rules you define.
            </p>
          </div>
          {variants && variants.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <ButtonWithPlus
                variant="primary"
                type="button"
                onClick={() => {
                  setEditingVariant(null);
                  setShowVariantForm(true);
                }}
              >
                New Variant
              </ButtonWithPlus>
            </div>
          )}
        </div>
        <div className="mt-10">
          {variants && variants.length > 0 ? (
            useTableView ? (
              <VariantTable
                variants={variants}
                onEdit={(variant) => {
                  setEditingVariant(variant);
                  setShowVariantForm(true);
                }}
                onDelete={(variant) => {
                  setDeletingVariant(variant);
                  setShowDeleteVariantModal(true);
                }}
              />
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {variants.map((variant) => (
                  <VariantCard
                    key={variant.key}
                    variant={variant}
                    onEdit={() => {
                      setEditingVariant(variant);
                      setShowVariantForm(true);
                    }}
                    onDelete={() => {
                      setDeletingVariant(variant);
                      setShowDeleteVariantModal(true);
                    }}
                  />
                ))}
              </div>
            )
          ) : (
            <Well>
              <SlidersHorizontalIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-4 dark:text-gray-200">
                No Variants Yet
              </h3>
              <Button
                variant="primary"
                aria-label="New Variant"
                onClick={(e) => {
                  e.preventDefault();
                  setEditingVariant(null);
                  setShowVariantForm(true);
                }}
              >
                Create Variant
              </Button>
            </Well>
          )}
        </div>
      </div>
    </>
  );
}
