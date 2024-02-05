import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { forwardRef } from 'react';
import * as Yup from 'yup';
import {
  useCreateNamespaceMutation,
  useUpdateNamespaceMutation
} from '~/app/namespaces/namespacesSlice';
import Button from '~/components/forms/buttons/Button';
import Input from '~/components/forms/Input';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { INamespace, INamespaceBase } from '~/types/Namespace';
import { stringAsKey } from '~/utils/helpers';

const namespaceValidationSchema = Yup.object({
  key: keyValidation,
  name: requiredValidation
});

type NamespaceFormProps = {
  setOpen: (open: boolean) => void;
  namespace?: INamespace;
  onSuccess: () => void;
};

const NamespaceForm = forwardRef((props: NamespaceFormProps, ref: any) => {
  const { setOpen, namespace, onSuccess } = props;

  const isNew = namespace === undefined;
  const title = isNew ? 'New Namespace' : 'Edit Namespace';
  const submitPhrase = isNew ? 'Create' : 'Update';

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const [createNamespace] = useCreateNamespaceMutation();
  const [updateNamespace] = useUpdateNamespaceMutation();

  const handleSubmit = async (values: INamespaceBase) => {
    if (isNew) {
      return createNamespace(values).unwrap();
    }
    return updateNamespace({ key: namespace.key, values }).unwrap();
  };

  return (
    <Formik
      initialValues={{
        key: namespace?.key || '',
        name: namespace?.name || '',
        description: namespace?.description || ''
      }}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess(
              `Successfully ${submitPhrase.toLocaleLowerCase()}d namespace.`
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
      validationSchema={namespaceValidationSchema}
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
                  <MoreInfo href="https://www.flipt.io/docs/concepts#namespaces">
                    Learn more about namespaces
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
                    htmlFor="name"
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Name
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Input
                    name="name"
                    id="name"
                    forwardRef={ref}
                    autoFocus={isNew}
                    onChange={(e) => {
                      // check if the name and key are currently in sync
                      // we do this so we don't override a custom key value
                      if (
                        isNew &&
                        (formik.values.key === '' ||
                          formik.values.key === stringAsKey(formik.values.name))
                      ) {
                        formik.setFieldValue(
                          'key',
                          stringAsKey(e.target.value)
                        );
                      }
                      formik.handleChange(e);
                    }}
                  />
                </div>
              </div>
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
                  <Input
                    name="key"
                    id="key"
                    disabled={!isNew}
                    onChange={(e) => {
                      const formatted = stringAsKey(e.target.value);
                      formik.setFieldValue('key', formatted);
                    }}
                  />
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

NamespaceForm.displayName = 'NamespaceForm';
export default NamespaceForm;
