import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik } from 'formik';
import { forwardRef } from 'react';
import * as Yup from 'yup';
import Button from '~/components/forms/Button';
import Input from '~/components/forms/Input';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { createToken } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { requiredValidation } from '~/data/validations';
import { IAuthTokenBase, IAuthTokenSecret } from '~/types/auth/Token';

type TokenFormProps = {
  setOpen: (open: boolean) => void;
  onSuccess: (token: IAuthTokenSecret) => void;
};

const TokenForm = forwardRef((props: TokenFormProps, ref: any) => {
  const { setOpen, onSuccess } = props;
  const { setError, clearError } = useError();

  const handleSubmit = async (values: IAuthTokenBase) => {
    createToken(values).then((resp) => {
      onSuccess(resp);
    });
  };

  const initialValues = {
    name: '',
    description: '',
    expires: ''
  };

  return (
    <Formik
      initialValues={initialValues}
      validationSchema={Yup.object({
        name: requiredValidation,
        description: requiredValidation
      })}
      onSubmit={(values, { setSubmitting }) => {
        let token: IAuthTokenBase = {
          name: values.name,
          description: values.description
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
        <Form className="flex h-full flex-col overflow-y-scroll bg-white shadow-xl">
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
            </div>
          </div>
          <div className="flex-shrink-0 border-t border-gray-200 px-4 py-5 sm:px-6">
            <div className="flex justify-end space-x-3">
              <Button
                onClick={() => {
                  setOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                primary
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
