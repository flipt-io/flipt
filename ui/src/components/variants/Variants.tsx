import { BracesIcon, BracketsIcon, SlidersHorizontalIcon, XIcon } from 'lucide-react';
import { useContext, useRef, useState } from 'react';

import { ButtonWithPlus } from '~/components/Button';
import Modal from '~/components/Modal';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import VariantForm from '~/components/variants/VariantForm';

import { IVariant } from '~/types/Variant';

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
              <span className="font-medium text-violet-500">
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
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
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
          ) : (
            <Well>
              <SlidersHorizontalIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-4 dark:text-gray-200">
                No Variants Yet
              </h3>
              <button
                aria-label="New Variant"
                onClick={(e) => {
                  e.preventDefault();
                  setEditingVariant(null);
                  setShowVariantForm(true);
                }}
                className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-violet-500 text-white hover:bg-violet-600 h-9 px-4 py-2"
              >
                Create Variant
              </button>
            </Well>
          )}
        </div>
      </div>
    </>
  );
}

function VariantCard({ variant, onEdit, onDelete }: { 
  variant: IVariant, 
  onEdit: () => void, 
  onDelete: () => void 
}) {
  // Check if variant has a non-empty attachment
  const hasAttachment = variant.attachment && 
    Object.keys(variant.attachment).length > 0;

  return (
    <div 
      className="group rounded-lg border border-gray-200 dark:border-gray-800 text-left text-sm transition-all hover:bg-gray-50 dark:hover:bg-gray-800 h-full flex flex-col cursor-pointer relative overflow-hidden shadow-sm hover:shadow"
      onClick={(e) => {
        e.preventDefault();
        onEdit();
      }}
    >
      <div className="flex flex-col p-4 h-full">
        {/* Header with variant icon and delete button */}
        <div className="flex items-center justify-between mb-5">
          <span className="rounded-md p-1.5 flex items-center space-x-2 justify-center bg-violet-100 text-violet-900 dark:bg-violet-900 dark:text-violet-100" title="Variant">
            <SlidersHorizontalIcon className="h-4 w-4" />
            <span className="text-xs">Variant</span>
          </span>
          
          <div className="flex items-center gap-2">
            <button
              onClick={(e) => {
                e.stopPropagation(); // Prevent card click from triggering
                onDelete();
              }}
              className="p-1.5 rounded-full text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
              aria-label="Delete variant"
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
        
        <div className="flex flex-col min-w-0 flex-1">
          {/* Simple label-value pairs format */}
          <div className="flex flex-col gap-4">
            {/* Key row */}
            <div className="flex items-baseline">
              <span className="text-gray-500 dark:text-gray-400 text-sm font-medium uppercase w-24">
                KEY:
              </span>
              <div className="flex items-center">
                <code className="text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded font-mono text-sm">
                  {variant.key}
                </code>
              </div>
            </div>
            
            {/* Name row (if present) */}
            {variant.name && (
              <div className="flex items-baseline">
                <span className="text-gray-500 dark:text-gray-400 text-sm font-medium uppercase w-24">
                  NAME:
                </span>
                <span className="text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-sm">
                  {variant.name}
                </span>
              </div>
            )}
          </div>
          
          {/* Description (if present) */}
          {variant.description && (
            <div className="mt-4 pt-3 border-t border-gray-100 dark:border-gray-800">
              <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                {variant.description}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
