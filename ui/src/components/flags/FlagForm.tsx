import { Form, Formik } from 'formik';
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
import { stringAsKey } from '~/utils/helpers';

type FlagFormProps = {
  flag?: IFlag;
  flagChanged?: () => void;
};

const flagTypes = [
  {
    id: 'VARIANT_FLAG_TYPE',
    name: FlagType.VARIANT_FLAG_TYPE
  },
  {
    id: 'BOOLEAN_FLAG_TYPE',
    name: FlagType.BOOLEAN_FLAG_TYPE
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
    type: flag?.type || ('VARIANT_FLAG_TYPE' as FlagType),
    enabled: flag?.enabled || false
  };

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
                    className="block text-sm font-medium text-gray-700"
                  >
                    Name
                  </label>
                  <Input
                    className="mt-1"
                    name="name"
                    id="name"
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
                    className="block text-sm font-medium text-gray-700"
                  >
                    Key
                  </label>
                  <Input
                    className="mt-1"
                    name="key"
                    id="key"
                    disabled={!isNew}
                    onChange={(e) => {
                      const formatted = stringAsKey(e.target.value);
                      formik.setFieldValue('key', formatted);
                    }}
                  />
                </div>
                <div className="col-span-3">
                  <label
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-700"
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
                              disabled={!isNew}
                              className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                              onChange={() => {
                                formik.setFieldValue(
                                  'type',
                                  type.id as FlagType
                                );
                              }}
                              checked={type.id === formik.values.type}
                              value={type.id}
                            />
                          </div>
                          <div className="ml-3 text-sm">
                            <label
                              htmlFor={type.id}
                              className="font-medium text-gray-700"
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
                      className="block text-sm font-medium text-gray-700"
                    >
                      Description
                    </label>
                    <span
                      className="text-xs text-gray-500"
                      id="description-optional"
                    >
                      Optional
                    </span>
                  </div>
                  <Input className="mt-1" name="description" id="description" />
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
