import { PlusIcon } from '@heroicons/react/24/outline';
import { useRef, useState } from 'react';
import { useSelector } from 'react-redux';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import EmptyState from '~/components/EmptyState';
import VariantForm from '~/components/flags/VariantForm';
import Button from '~/components/forms/buttons/Button';
import Modal from '~/components/Modal';
import DeletePanel from '~/components/panels/DeletePanel';
import Slideover from '~/components/Slideover';
import { deleteVariant } from '~/data/api';
import { IFlag } from '~/types/Flag';
import { IVariant } from '~/types/Variant';

type VariantsProps = {
  flag: IFlag;
  flagChanged: () => void;
};

export default function Variants(props: VariantsProps) {
  const { flag, flagChanged } = props;

  const [showVariantForm, setShowVariantForm] = useState<boolean>(false);
  const [editingVariant, setEditingVariant] = useState<IVariant | null>(null);
  const [showDeleteVariantModal, setShowDeleteVariantModal] =
    useState<boolean>(false);
  const [deletingVariant, setDeletingVariant] = useState<IVariant | null>(null);

  const variantFormRef = useRef(null);

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

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
            flagChanged();
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
          handleDelete={
            () =>
              deleteVariant(namespace.key, flag.key, deletingVariant?.id ?? '') // TODO: Determine impact of blank ID param
          }
          onSuccess={() => {
            flagChanged();
          }}
        />
      </Modal>

      {/* variants */}
      <div className="mt-10">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-lg font-medium leading-6 text-gray-900">
              Variants
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Return different values based on rules you define
            </p>
          </div>
          {flag.variants && flag.variants.length > 0 && (
            <div className="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
              <Button
                primary
                type="button"
                disabled={readOnly}
                title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
                onClick={() => {
                  setEditingVariant(null);
                  setShowVariantForm(true);
                }}
              >
                <PlusIcon
                  className="-ml-1.5 mr-1 h-5 w-5 text-white"
                  aria-hidden="true"
                />
                <span>New Variant</span>
              </Button>
            </div>
          )}
        </div>
        <div className="my-10">
          {flag.variants && flag.variants.length > 0 ? (
            <table className="min-w-full divide-y divide-gray-300">
              <thead>
                <tr>
                  <th
                    scope="col"
                    className="pb-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6"
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
                  <th scope="col" className="relative pb-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Edit</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {flag.variants.map((variant) => (
                  <tr key={variant.key}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm text-gray-600 sm:pl-6">
                      {variant.key}
                    </td>
                    <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 sm:table-cell">
                      {variant.name}
                    </td>
                    <td className="hidden truncate whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                      {variant.description}
                    </td>
                    <td className="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
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
