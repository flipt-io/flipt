import { PlusIcon } from '@heroicons/react/24/outline';
import { useRef, useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import DeletePanel from '~/components/DeletePanel';
import EmptyState from '~/components/EmptyState';
import FlagForm from '~/components/flags/FlagForm';
import VariantForm from '~/components/flags/VariantForm';
import Button from '~/components/forms/Button';
import Modal from '~/components/Modal';
import MoreInfo from '~/components/MoreInfo';
import Slideover from '~/components/Slideover';
import { deleteVariant } from '~/data/api';
import useNamespace from '~/data/hooks/namespace';
import { IVariant } from '~/types/Variant';
import { FlagProps } from './FlagProps';

export default function EditFlag() {
  const { flag, onFlagChange } = useOutletContext<FlagProps>();

  const [showVariantForm, setShowVariantForm] = useState<boolean>(false);
  const [editingVariant, setEditingVariant] = useState<IVariant | null>(null);
  const [showDeleteVariantModal, setShowDeleteVariantModal] =
    useState<boolean>(false);
  const [deletingVariant, setDeletingVariant] = useState<IVariant | null>(null);

  const variantFormRef = useRef(null);

  const { currentNamespace } = useNamespace();

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
            onFlagChange();
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
              deleteVariant(
                currentNamespace.key,
                flag.key,
                deletingVariant?.id ?? ''
              ) // TODO: Determine impact of blank ID param
          }
          onSuccess={() => {
            onFlagChange();
          }}
        />
      </Modal>

      <div className="flex flex-col">
        {/* flag details */}
        <div className="my-10">
          <div className="md:grid md:grid-cols-3 md:gap-6">
            <div className="md:col-span-1">
              <p className="mt-1 text-sm text-gray-500">
                Basic information about the flag and its status.
              </p>
              <MoreInfo
                className="mt-5"
                href="https://www.flipt.io/docs/concepts#flags"
              >
                Learn more about flags
              </MoreInfo>
            </div>
            <div className="mt-5 md:col-span-2 md:mt-0">
              <FlagForm flag={flag} flagChanged={onFlagChange} />
            </div>
          </div>
        </div>

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
                    <th
                      scope="col"
                      className="relative pb-3.5 pl-3 pr-4 sm:pr-6"
                    >
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
                      <td className="hidden whitespace-nowrap px-3 py-4 text-sm text-gray-500 lg:table-cell">
                        {variant.description}
                      </td>
                      <td className="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
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
                        |
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
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <EmptyState
                text="New Variant"
                onClick={() => {
                  setEditingVariant(null);
                  setShowVariantForm(true);
                }}
              />
            )}
          </div>
        </div>
      </div>
    </>
  );
}
