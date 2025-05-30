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
            <p className="mt-1 text-sm text-gray-500">
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
            <table className="min-w-full divide-y divide-gray-300">
              <thead>
                <tr>
                  <th
                    scope="col"
                    className="pr-3 pb-3.5 pl-4 text-left text-sm font-semibold text-gray-900 sm:pl-6"
                  >
                    Key
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 sm:table-cell"
                  >
                    Name
                  </th>
                  <th
                    scope="col"
                    className="hidden px-3 pb-3.5 text-left text-sm font-semibold text-gray-900 lg:table-cell"
                  >
                    Description
                  </th>
                  <th scope="col" className="relative pr-4 pb-3.5 pl-3 sm:pr-6">
                    <span className="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {flag.variants.map((variant) => (
                  <tr key={variant.key}>
                    <td className="py-4 pr-3 pl-4 text-sm whitespace-nowrap text-gray-600 sm:pl-6">
                      {variant.key}
                    </td>
                    <td className="hidden px-3 py-4 text-sm whitespace-nowrap text-gray-500 sm:table-cell">
                      {variant.name}
                    </td>
                    <td className="hidden truncate px-3 py-4 text-sm whitespace-nowrap text-gray-500 lg:table-cell">
                      {variant.description}
                    </td>
                    <td className="py-4 pr-4 pl-3 text-right text-sm font-medium whitespace-nowrap sm:pr-6">
                      {!readOnly && (
                        <>
                          <a
                            href="#"
                            className="pr-2 text-violet-600 hover:text-violet-900"
                            onClick={(e) => {
                              e.preventDefault();
                              setEditingVariant(variant);
                              setShowVariantForm(true);
                            }}
                          >
                            Edit
                            <span className="sr-only">,{variant.key}</span>
                          </a>
                          <span aria-hidden="true"> | </span>
                          <a
                            href="#"
                            className="pl-2 text-violet-600 hover:text-violet-900"
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
