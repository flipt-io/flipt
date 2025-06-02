import { SlidersHorizontalIcon } from 'lucide-react';
import { useContext, useRef, useState } from 'react';

import { Button, ButtonWithPlus } from '~/components/Button';
import Slideover from '~/components/Slideover';
import Well from '~/components/Well';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import DeletePanel from '~/components/panels/DeletePanel';
import VariantForm from '~/components/variants/VariantForm';
import VariantTable from '~/components/variants/VariantTable';

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
      <DeletePanel
        open={showDeleteVariantModal}
        panelMessage={
          <>
            Are you sure you want to delete the variant{' '}
            <span className="font-medium text-brand">
              {deletingVariant?.key}
            </span>
            ?
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

      {/* variants */}
      <div className="mt-2 min-w-full">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="mt-1 text-sm text-muted-foreground">
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
            <Well>
              <SlidersHorizontalIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-4">
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
