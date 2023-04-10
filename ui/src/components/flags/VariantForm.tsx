import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { forwardRef } from 'react';
import * as Yup from 'yup';
import Button from '~/components/forms/Button';
import Input from '~/components/forms/Input';
import TextArea from '~/components/forms/TextArea';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { createVariant, updateVariant } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { jsonValidation, keyValidation } from '~/data/validations';
import { IVariant, IVariantBase } from '~/types/Variant';

type VariantFormProps = {
  setOpen: (open: boolean) => void;
  flagKey: string;
  variant?: IVariant;
  onSuccess: () => void;
};

const VariantForm = forwardRef((props: VariantFormProps, ref: any) => {
  const { setOpen, flagKey, variant, onSuccess } = props;

  const isNew = variant === undefined;
  const title = isNew ? 'New Variant' : 'Edit Variant';
  const submitPhrase = isNew ? 'Create' : 'Update';

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const handleSubmit = async (values: IVariantBase) => {
    if (isNew) {
      return createVariant(currentNamespace.key, flagKey, values);
    }

    return updateVariant(currentNamespace.key, flagKey, variant?.id, values);
  };

  return (
    <Formik
      initialValues={{
        key: variant?.key || '',
        name: variant?.name || '',
        description: variant?.description || '',
        attachment: variant?.attachment || ''
      }}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess(
              `Successfully ${submitPhrase.toLocaleLowerCase()}d variant.`
            );
            onSuccess();
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
      validationSchema={Yup.object({
        key: keyValidation,
        attachment: jsonValidation
      })}
    >
      {(formik) => (
        <Form className="flex h-full flex-col overflow-y-scroll bg-white shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-lg font-medium text-gray-900">
                    {title}
                  </Dialog.Title>
                  <MoreInfo href="https://www.flipt.io/docs/concepts#variants">
                    Learn more about variants
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500"
                    onClick={() => setOpen(false)}
                  >
                    <span className="sr-only">Close panel</span>
                    <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>
            <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 sm:py-0">
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="key"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Key
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Input name="key" id="key" forwardRef={ref} />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="name"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Name
                  </label>
                  <span className="text-xs text-gray-400" id="name-optional">
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <Input name="name" id="name" />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="description"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Description
                  </label>
                  <span
                    className="text-xs text-gray-400"
                    id="description-optional"
                  >
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <Input name="description" id="description" />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="attachment"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Attachment
                  </label>
                  <span
                    className="text-xs text-gray-400"
                    id="attachment-optional"
                  >
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <TextArea name="attachment" id="attachment" />
                </div>
              </div>
            </div>
          </div>
          <div className="flex-shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button onClick={() => setOpen(false)}>Cancel</Button>
              <Button
                primary
                className="min-w-[80px]"
                type="submit"
                disabled={
                  !(formik.dirty && formik.isValid && !formik.isSubmitting)
                }
              >
                {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
              </Button>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
});

VariantForm.displayName = 'VariantForm';
export default VariantForm;
