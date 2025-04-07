import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { forwardRef, useState } from 'react';
import { useSelector } from 'react-redux';
import * as Yup from 'yup';
import { selectNamespaces } from '~/app/namespaces/namespacesSlice';
import { useCreateTokenMutation } from '~/app/tokens/tokensApi';
import { Button } from '~/components/Button';
import Input from '~/components/forms/Input';
import Listbox from '~/components/forms/Listbox';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { SelectableNamespace } from '~/components/namespaces/NamespaceListbox';
import { useError } from '~/data/hooks/error';
import { requiredValidation } from '~/data/validations';
import { IAuthTokenBase, IAuthTokenSecret } from '~/types/auth/Token';
import { INamespace } from '~/types/Namespace';

const tokenValidationSchema = Yup.object({
  name: requiredValidation
});

type TokenFormProps = {
  setOpen: (open: boolean) => void;
  onSuccess: (token: IAuthTokenSecret) => void;
};

const TokenForm = forwardRef((props: TokenFormProps, ref: any) => {
  const { setOpen, onSuccess } = props;
  const { setError, clearError } = useError();
  const [createToken] = useCreateTokenMutation();

  const handleSubmit = async (values: IAuthTokenBase) => {
    createToken(values)
      .unwrap()
      .then((resp) => {
        onSuccess(resp);
      });
  };

  const namespaces = useSelector(selectNamespaces) as INamespace[];

  const [namespaceScoped, setNamespaceScoped] = useState<boolean>(false);

  const [selectedNamespace, setSelectedNamespace] =
    useState<SelectableNamespace>({
      ...namespaces[0],
      displayValue: namespaces[0].name
    });

  const initialValues = {
    name: '',
    description: '',
    namespaceKey: '',
    expires: ''
  };

  return (
    <Formik
      initialValues={initialValues}
      validationSchema={tokenValidationSchema}
      onSubmit={(values, { setSubmitting }) => {
        let token: IAuthTokenBase = {
          name: values.name,
          description: values.description?.trim() || '',
          namespaceKey: values.namespaceKey || ''
        };

        // parse expires into UTC date
        if (values.expires) {
          let d = new Date(values.expires);
          d.setHours(24, 0, 0, 0); // set to 24:00:00 localtime (nearest midnight in future)
          token.expiresAt = d.toISOString();
        }

        handleSubmit(token)
          .then(() => {
            clearError();
          })
          .catch((err) => {
            setError(err.message);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
    >
      {(formik) => (
        <Form className="bg-background flex h-full flex-col overflow-y-scroll shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-lg font-medium text-gray-900">
                    New Token
                  </Dialog.Title>
                  <MoreInfo href="https://www.flipt.io/docs/authentication/methods#static-token">
                    Learn more about static tokens
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500"
                    onClick={() => {
                      setOpen(false);
                    }}
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
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Name
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Input name="name" id="name" forwardRef={ref} />
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
                    htmlFor="expires"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Expires On
                  </label>
                  <span className="text-xs text-gray-400" id="expires-optional">
                    Optional
                  </span>
                </div>
                <div className="sm:col-span-2">
                  <Input
                    type="date"
                    id="expires"
                    name="expires"
                    min={new Date().toISOString().split('T')[0]}
                  />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="expires"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Namespace
                  </label>
                  <span className="text-xs text-gray-400" id="expires-optional">
                    Optional
                  </span>
                </div>
                <div className="space-y-4 sm:col-span-2">
                  <div className="relative flex items-start">
                    <div className="flex h-6 items-center">
                      <input
                        id="namespaced"
                        name="namespaced"
                        type="checkbox"
                        className="h-4 w-4 rounded-sm border-gray-300 text-violet-600 focus:ring-violet-600"
                        onChange={(e) => {
                          setNamespaceScoped(e.target.checked);
                          formik.setFieldValue(
                            'namespaceKey',
                            e.target.checked ? selectedNamespace.key : ''
                          );
                        }}
                      />
                    </div>
                    <div className="ml-3 text-sm leading-6">
                      <label
                        htmlFor="namespaced"
                        className="font-medium text-gray-700"
                      >
                        Scope this token to a single namespace
                      </label>
                    </div>
                  </div>
                  {namespaceScoped && (
                    <Listbox<SelectableNamespace>
                      id="tokenNamespace"
                      name="namespaceKey"
                      values={namespaces.map((n) => ({
                        ...n,
                        displayValue: n.name
                      }))}
                      selected={{
                        ...selectedNamespace,
                        displayValue: selectedNamespace?.name || ''
                      }}
                      setSelected={(v) => {
                        setSelectedNamespace(v);
                        formik.setFieldValue('namespaceKey', v.key);
                      }}
                    />
                  )}
                </div>
              </div>
            </div>
          </div>
          <div className="shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button
                onClick={() => {
                  setOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                type="submit"
                className="min-w-[80px]"
                disabled={
                  !(formik.dirty && formik.isValid && !formik.isSubmitting)
                }
              >
                {formik.isSubmitting ? <Loading isPrimary /> : 'Create'}
              </Button>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
});

TokenForm.displayName = 'TokenForm';
export default TokenForm;
