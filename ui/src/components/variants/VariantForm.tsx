import * as Dialog from '@radix-ui/react-dialog';
import { Form, Formik } from 'formik';
import { XIcon } from 'lucide-react';
import { forwardRef, useContext } from 'react';
import * as Yup from 'yup';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { FlagFormContext } from '~/components/flags/FlagFormContext';
import Input from '~/components/forms/Input';
import { JsonTextArea } from '~/components/forms/JsonTextArea';

import { IVariant } from '~/types/Variant';

import { useError } from '~/data/hooks/error';
import { keyWithDotValidation } from '~/data/validations';

const validationSchema = Yup.object({
  key: keyWithDotValidation
});

type VariantFormProps = {
  setOpen: (open: boolean) => void;
  variant?: IVariant | null;
  onSuccess: () => void;
};

const VariantForm = forwardRef((props: VariantFormProps, ref: any) => {
  const { setOpen, variant, onSuccess } = props;

  const isNew = variant === null;
  const title = isNew ? 'New Variant' : 'Edit Variant';
  const submitPhrase = isNew ? 'Add' : 'Done';

  const { setError, clearError } = useError();

  const { updateVariant, createVariant } = useContext(FlagFormContext);

  const handleSubmit = async (values: IVariant) => {
    if (isNew) {
      createVariant(values);
      return;
    }
    updateVariant(values);
  };

  return (
    <Formik
      validateOnChange
      validate={(values: IVariant) => {
        !isNew && handleSubmit(values);
      }}
      initialValues={{
        key: variant?.key || '',
        name: variant?.name || '',
        description: variant?.description || '',
        attachment: variant?.attachment || {}
      }}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            onSuccess();
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
      validationSchema={validationSchema}
    >
      {(formik) => (
        <Form className="flex h-full flex-col overflow-y-scroll bg-background dark:bg-gray-900 shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 dark:bg-gray-800 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    {title}
                  </Dialog.Title>
                  <MoreInfo href="https://docs.flipt.io/v2/concepts#variants">
                    Learn more about variants
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
                    onClick={() => setOpen(false)}
                  >
                    <span className="sr-only">Close panel</span>
                    <XIcon className="h-6 w-6" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>
            <div className="space-y-6 py-6 sm:space-y-0 sm:divide-y sm:divide-gray-200 dark:sm:divide-gray-700 sm:py-0">
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="key"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Key
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Input
                    name="key"
                    id="key"
                    forwardRef={ref}
                    disabled={!isNew}
                  />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="name"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Name
                  </label>
                  <span
                    className="text-xs text-gray-400 dark:text-gray-500"
                    id="name-optional"
                  >
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
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Description
                  </label>
                  <span
                    className="text-xs text-gray-400 dark:text-gray-500"
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
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Attachment
                  </label>
                  <span
                    className="text-xs text-gray-400 dark:text-gray-500"
                    id="attachment-optional"
                  >
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <JsonTextArea name="attachment" id="attachment" />
                </div>
              </div>
            </div>
          </div>
          <div className="shrink-0 border-t border-gray-200 dark:border-gray-700 px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button variant="secondary" onClick={() => setOpen(false)}>
                Cancel
              </Button>
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
