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
    <div className="relative flex flex-col rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-950 overflow-hidden shadow-sm hover:shadow-md group">
      <div className="flex justify-between items-start p-2">
        <span className="p-1.5 rounded-md bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-100">
          <SlidersHorizontalIcon className="h-4 w-4" />
        </span>

        <div className="relative group/delete">
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onDelete();
            }}
            className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            aria-label="Delete variant"
          >
            <XIcon className="h-4 w-4" />
          </button>
          <div className="absolute hidden group-hover/delete:block -bottom-1 right-0 transform translate-y-full bg-gray-800 text-white text-xs px-2 py-1 rounded whitespace-nowrap z-10">
            Delete Variant
          </div>
        </div>
      </div>

      <div className="flex-1 p-4 pt-0" onClick={onEdit}>
        <div className="grid grid-cols-[120px_1fr] gap-y-3 items-start">
          <span className="text-sm font-medium uppercase text-gray-500 dark:text-gray-400 pt-1">
            KEY:
          </span>
          <code className="text-sm font-mono text-gray-900 dark:text-gray-100 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded truncate block overflow-hidden">
            {variant.key}
          </code>

          {variant.name && (
            <>
              <span className="text-sm font-medium uppercase text-gray-500 dark:text-gray-400 pt-1">
                NAME:
              </span>
              <span className="text-sm text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded truncate">
                {variant.name}
              </span>
            </>
          )}

          {hasAttachment && (
            <>
              <span className="text-sm font-medium uppercase text-gray-500 dark:text-gray-400 pt-1">
                ATTACHMENT:
              </span>
              <span className="text-sm text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded">
                {Object.keys(variant.attachment || {}).length} fields
              </span>
            </>
          )}

          {variant.description && (
            <>
              <span className="text-sm font-medium uppercase text-gray-500 dark:text-gray-400 pt-1">
                DESCRIPTION:
              </span>
              <div className="text-sm text-gray-700 dark:text-gray-300 max-h-20 overflow-y-auto">
                <p className="break-words">{variant.description}</p>
              </div>
            </>
          )}
        </div>
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
                <td className="px-3 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100 max-w-[200px]">
                  <div className="truncate">
                    {variant.name || (
                      <span className="text-gray-400 dark:text-gray-500">
                        —
                      </span>
                    )}
                  </div>
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
                <td className="px-3 py-4 text-sm text-gray-700 dark:text-gray-300 max-w-[300px]">
                  {variant.description ? (
                    <p className="line-clamp-1">{variant.description}</p>
                  ) : (
                    <span className="text-gray-400 dark:text-gray-500">—</span>
                  )}
                </td>
                <td className="px-3 py-4 whitespace-nowrap text-right text-sm font-medium">
                  <div className="relative group/delete inline-block">
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(variant);
                      }}
                      className="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                      aria-label="Delete variant"
                    >
                      <XIcon className="h-4 w-4" />
                    </button>
                    <div className="absolute hidden group-hover/delete:block -bottom-1 right-0 transform translate-y-full bg-gray-800 text-white text-xs px-2 py-1 rounded whitespace-nowrap z-10">
                      Delete Variant
                    </div>
                  </div>
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
    // Default auto behavior - table for more than 4 items
    return variants.length > 4;
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
              <div className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-3">
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
