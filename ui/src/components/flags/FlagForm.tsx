import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/react/20/solid';
import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import * as Yup from 'yup';
import { selectReadonly } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import Button from '~/components/forms/buttons/Button';
import Input from '~/components/forms/Input';
import Toggle from '~/components/forms/Toggle';
import Loading from '~/components/Loading';
import { createFlag, updateFlag } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { FlagType, IFlag, IFlagBase } from '~/types/Flag';
import { classNames, copyTextToClipboard, stringAsKey } from '~/utils/helpers';

type FlagFormProps = {
  flag?: IFlag;
  flagChanged?: () => void;
};

const flagTypes = [
  {
    id: FlagType.VARIANT,
    name: 'Variant'
  },
  {
    id: FlagType.BOOLEAN,
    name: 'Boolean'
  }
];

export default function FlagForm(props: FlagFormProps) {
  const { flag, flagChanged } = props;

  const isNew = flag === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';

  const navigate = useNavigate();
  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const namespace = useSelector(selectCurrentNamespace);
  const readOnly = useSelector(selectReadonly);

  const handleSubmit = (values: IFlagBase) => {
    if (isNew) {
      return createFlag(namespace.key, values);
    }
    return updateFlag(namespace.key, flag?.key, values);
  };

  const initialValues: IFlagBase = {
    key: flag?.key || '',
    name: flag?.name || '',
    description: flag?.description || '',
    type: flag?.type || FlagType.VARIANT,
    enabled: flag?.enabled || false
  };

  const [keyCopied, setKeyCopied] = useState(false);

  return (
    <Formik
      enableReinitialize
      initialValues={initialValues}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess(
              `Successfully ${submitPhrase.toLocaleLowerCase()}d flag`
            );
            if (isNew) {
              navigate(`/namespaces/${namespace.key}/flags/${values.key}`);
              return;
            }

            flagChanged && flagChanged();
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
        name: requiredValidation
      })}
    >
      {(formik) => {
        const { enabled } = formik.values;
        return (
          <Form className="px-1 sm:overflow-hidden sm:rounded-md">
            <div className="space-y-6">
              <div className="grid grid-cols-3 gap-6">
                <div className="col-span-3 md:col-span-2">
                  <Toggle
                    id="enabled"
                    name="enabled"
                    label="Enabled"
                    disabled={readOnly}
                    checked={enabled}
                    onChange={(e) => {
                      formik.setFieldValue('enabled', e);
                    }}
                  />
                </div>
                <div className="col-span-2">
                  <label
                    htmlFor="name"
                    className="text-gray-700 block text-sm font-medium"
                  >
                    Name
                  </label>
                  <Input
                    className="mt-1"
                    name="name"
                    id="name"
                    disabled={readOnly}
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
                <div className="col-span-2">
                  <label
                    htmlFor="key"
                    className="text-gray-700 block text-sm font-medium"
                  >
                    Key
                  </label>
                  <div
                    className={classNames(
                      isNew ? '' : 'flex items-center justify-between'
                    )}
                  >
                    <Input
                      className={classNames(isNew ? 'mt-1' : 'mt-1 md:mr-2')}
                      name="key"
                      id="key"
                      disabled={!isNew || readOnly}
                      onChange={(e) => {
                        const formatted = stringAsKey(e.target.value);
                        formik.setFieldValue('key', formatted);
                      }}
                    />
                    {!isNew && (
                      <button
                        aria-label="Copy"
                        title="Copy to Clipboard"
                        className="hidden md:block"
                        onClick={(e) => {
                          e.preventDefault();
                          copyTextToClipboard(flag?.key || '');
                          setKeyCopied(true);
                          setTimeout(() => {
                            setKeyCopied(false);
                          }, 2000);
                        }}
                      >
                        <CheckIcon
                          className={classNames(
                            'nightwind-prevent text-green-400 absolute m-auto h-5 w-5 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-white',
                            keyCopied
                              ? 'visible opacity-100'
                              : 'invisible opacity-0'
                          )}
                        />
                        <ClipboardDocumentIcon
                          className={classNames(
                            'text-gray-400 m-auto h-5 w-5 justify-center align-middle transition-opacity duration-300 ease-in-out hover:text-white dark:hover:text-gray-500',
                            keyCopied
                              ? 'invisible opacity-0'
                              : 'visible opacity-100'
                          )}
                        />
                      </button>
                    )}
                  </div>
                </div>
                <div className="col-span-3">
                  <label
                    htmlFor="type"
                    className="text-gray-700 block text-sm font-medium"
                  >
                    Type
                  </label>
                  <fieldset className="mt-2">
                    <legend className="sr-only">Type</legend>
                    <div className="flex flex-row space-x-5">
                      {flagTypes.map((type) => (
                        <div
                          key={type.id}
                          className="relative flex items-start"
                        >
                          <div className="flex h-5 items-center">
                            <input
                              id={type.id}
                              name="type"
                              type="radio"
                              disabled={!isNew || readOnly}
                              className="text-violet-400 border-gray-300 h-4 w-4 focus:ring-violet-400"
                              onChange={() => {
                                formik.setFieldValue('type', type.id);
                              }}
                              checked={type.id === formik.values.type}
                              value={type.id}
                            />
                          </div>
                          <div className="ml-3 text-sm">
                            <label
                              htmlFor={type.id}
                              className="text-gray-700 font-medium"
                            >
                              {type.name}
                            </label>
                          </div>
                        </div>
                      ))}
                    </div>
                  </fieldset>
                </div>
                <div className="col-span-3">
                  <div className="flex justify-between">
                    <label
                      htmlFor="description"
                      className="text-gray-700 block text-sm font-medium"
                    >
                      Description
                    </label>
                    <span
                      className="text-gray-500 text-xs"
                      id="description-optional"
                    >
                      Optional
                    </span>
                  </div>
                  <Input
                    className="mt-1"
                    name="description"
                    id="description"
                    disabled={readOnly}
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button type="button" onClick={() => navigate(-1)}>
                  Cancel
                </Button>
                <Button
                  primary
                  className="ml-3 min-w-[80px]"
                  type="submit"
                  title={readOnly ? 'Not allowed in Read-Only mode' : undefined}
                  disabled={
                    !(formik.dirty && formik.isValid && !formik.isSubmitting) ||
                    readOnly
                  }
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                </Button>
              </div>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
}
