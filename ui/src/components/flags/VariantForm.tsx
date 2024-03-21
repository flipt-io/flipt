import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { forwardRef } from 'react';
import { useSelector } from 'react-redux';
import * as Yup from 'yup';
import {
  useCreateVariantMutation,
  useUpdateVariantMutation
} from '~/app/flags/flagsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import Button from '~/components/forms/buttons/Button';
import Input from '~/components/forms/Input';
import TextArea from '~/components/forms/TextArea';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { jsonValidation, keyWithDotValidation } from '~/data/validations';
import { IVariant, IVariantBase } from '~/types/Variant';

const variantValidationSchema = Yup.object({
  key: keyWithDotValidation,
  attachment: jsonValidation
});

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

  const namespace = useSelector(selectCurrentNamespace);

  const [createVariant] = useCreateVariantMutation();
  const [updateVariant] = useUpdateVariantMutation();

  const handleSubmit = async (values: IVariantBase) => {
    if (isNew) {
      return createVariant({
        namespaceKey: namespace.key,
        flagKey: flagKey,
        values: values
      }).unwrap();
    }

    return updateVariant({
      namespaceKey: namespace.key,
      flagKey: flagKey,
      variantId: variant?.id,
      values: values
    }).unwrap();
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
      validationSchema={variantValidationSchema}
    >
      {(formik) => (
        <Form className="bg-white flex h-full flex-col overflow-y-scroll shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-gray-900 text-lg font-medium">
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Name
                  </label>
                  <span className="text-gray-400 text-xs" id="name-optional">
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Description
                  </label>
                  <span
                    className="text-gray-400 text-xs"
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Attachment
                  </label>
                  <span
                    className="text-gray-400 text-xs"
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
          <div className="border-gray-200 flex-shrink-0 border-t px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button onClick={() => setOpen(false)}>Cancel</Button>
              <Button
                variant="primary"
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
