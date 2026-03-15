import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { useDeleteVariantMutation } from '~/app/flags/flagsApi';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import EmptyState from '~/components/EmptyState';
import VariantForm from '~/components/variants/forms/VariantForm';
import { ButtonWithPlus } from '~/components/Button';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import Slideover from '~/components/Slideover';
import { IFlag } from '~/types/Flag';
import { IVariant } from '~/types/Variant';

type VariantsProps = {
  flag: IFlag;
};

export default function Variants({ flag }: VariantsProps) {
  const [showVariantForm, setShowVariantForm] = useState<boolean>(false);
  const [editingVariant, setEditingVariant] = useState<IVariant | null>(null);
  const [showDeleteVariantModal, setShowDeleteVariantModal] =
    useState<boolean>(false);
  const [deletingVariant, setDeletingVariant] = useState<IVariant | null>(null);

  const variantFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const [deleteVariant] = useDeleteVariantMutation();

  return (
    <>
      {/* variant edit form */}
      <Slideover
        open={showVariantForm}
        setOpen={setShowVariantForm}
        ref={variantFormRef}
        title={editingVariant ? 'Edit Variant' : 'New Variant'}
      >
        <VariantForm
          ref={variantFormRef}
          flagKey={flag.key}
          variant={editingVariant || undefined}
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
          handleDelete={() =>
            deleteVariant({
              namespaceKey: namespace.key,
              flagKey: flag.key,
              variantId: deletingVariant?.id ?? ''
            }).unwrap()
          }
        />
      </Modal>

      {/* variants */}
      <div className="mt-2">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <p className="text-muted-foreground mt-1 text-sm">
              Return different values based on rules you define.
            </p>
          </div>
          {flag.variants && flag.variants.length > 0 && (
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
              <ButtonWithPlus
                variant="primary"
                type="button"
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
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
          {flag.variants && flag.variants.length > 0 ? (
            <table className="divide-border min-w-full divide-y">
              <thead>
                <tr>
                  <th
                    scope="col"
                    className="text-secondary-foreground pr-3 pb-3.5 pl-4 text-left text-sm font-semibold sm:pl-6"
                  >
                    Key
                  </th>
                  <th
                    scope="col"
                    className="text-secondary-foreground hidden px-3 pb-3.5 text-left text-sm font-semibold sm:table-cell"
                  >
                    Name
                  </th>
                  <th
                    scope="col"
                    className="text-secondary-foreground hidden px-3 pb-3.5 text-left text-sm font-semibold lg:table-cell"
                  >
                    Description
                  </th>
                  <th scope="col" className="relative pr-4 pb-3.5 pl-3 sm:pr-6">
                    <span className="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-border divide-y">
                {flag.variants.map((variant) => (
                  <tr key={variant.key}>
                    <td className="text-secondary-foreground py-4 pr-3 pl-4 text-sm whitespace-nowrap sm:pl-6">
                      {variant.key}
                    </td>
                    <td className="text-muted-foreground hidden px-3 py-4 text-sm whitespace-nowrap sm:table-cell">
                      {variant.name}
                    </td>
                    <td className="text-muted-foreground hidden truncate px-3 py-4 text-sm whitespace-nowrap lg:table-cell">
                      {variant.description}
                    </td>
                    <td className="py-4 pr-4 pl-3 text-right text-sm font-medium whitespace-nowrap sm:pr-6">
                      {!readOnly && (
                        <>
                          <a
                            href="#"
                            className="text-brand/80 pr-2"
                            onClick={(e) => {
                              e.preventDefault();
                              setEditingVariant(variant);
                              setShowVariantForm(true);
                            }}
                          >
                            Edit
                            <span className="sr-only">,{variant.key}</span>
                          </a>
                          <span
                            aria-hidden="true"
                            className="text-muted-foreground"
                          >
                            {' '}
                            |{' '}
                          </span>
                          <a
                            href="#"
                            className="text-brand/80 pl-2"
                            onClick={(e) => {
                              e.preventDefault();
                              setDeletingVariant(variant);
                              setShowDeleteVariantModal(true);
                            }}
                          >
                            Delete
                            <span className="sr-only">,{variant.key}</span>
                          </a>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <EmptyState
              text="New Variant"
              disabled={readOnly}
              onClick={() => {
                setEditingVariant(null);
                setShowVariantForm(true);
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
