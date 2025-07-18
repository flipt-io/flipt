import * as Dialog from '@radix-ui/react-dialog';
import { Form, Formik } from 'formik';
import { XIcon } from 'lucide-react';
import { forwardRef } from 'react';
import { useSelector } from 'react-redux';
import * as Yup from 'yup';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import {
  useCreateNamespaceMutation,
  useUpdateNamespaceMutation
} from '~/app/namespaces/namespacesApi';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import Input from '~/components/forms/Input';

import { INamespace } from '~/types/Namespace';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { getRevision, stringAsKey } from '~/utils/helpers';

const namespaceValidationSchema = Yup.object({
  key: keyValidation,
  name: requiredValidation
});

type NamespaceFormProps = {
  setOpen: (open: boolean) => void;
  namespace?: INamespace | null;
  onSuccess: () => void;
};

const NamespaceForm = forwardRef((props: NamespaceFormProps, ref: any) => {
  const { setOpen, namespace, onSuccess } = props;

  const isNew = namespace === null;
  const title = isNew ? 'New Namespace' : 'Edit Namespace';
  const submitPhrase = isNew ? 'Create' : 'Update';

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const environment = useSelector(selectCurrentEnvironment);
  const revision = getRevision();

  const [createNamespace] = useCreateNamespaceMutation();
  const [updateNamespace] = useUpdateNamespaceMutation();

  const handleSubmit = async (values: INamespace) => {
    if (isNew) {
      return createNamespace({
        environmentKey: environment.key,
        values,
        revision
      }).unwrap();
    }
    return updateNamespace({
      environmentKey: environment.key,
      values,
      revision
    }).unwrap();
  };

  return (
    <Formik
      initialValues={{
        key: namespace?.key || '',
        name: namespace?.name || '',
        description: namespace?.description || '',
        protected: namespace?.protected || false
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
        <Form className="flex h-full flex-col overflow-y-scroll bg-background dark:bg-gray-900 shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 dark:bg-gray-800 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-lg font-medium text-gray-900 dark:text-gray-100">
                    {title}
                  </Dialog.Title>
                  <MoreInfo href="https://docs.flipt.io/v2/concepts#namespaces">
                    Learn more about namespaces
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-300"
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
                    htmlFor="name"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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

NamespaceForm.displayName = 'NamespaceForm';
export default NamespaceForm;
