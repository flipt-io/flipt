import { Form, Formik } from 'formik';
import { useNavigate } from 'react-router-dom';
import * as Yup from 'yup';
import Loading from '~/components//Loading';
import Button from '~/components/forms/Button';
import Input from '~/components/forms/Input';
import Toggle from '~/components/forms/Toggle';
import { createFlag, updateFlag } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { IFlag, IFlagBase } from '~/types/Flag';
import { stringAsKey } from '~/utils/helpers';

type FlagFormProps = {
  flag?: IFlag;
  flagChanged?: () => void;
};

export default function FlagForm(props: FlagFormProps) {
  const { flag, flagChanged } = props;

  const isNew = flag === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';

  const navigate = useNavigate();
  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const handleSubmit = (values: IFlagBase) => {
    if (isNew) {
      return createFlag(currentNamespace.key, values);
    }
    return updateFlag(currentNamespace.key, flag?.key, values);
  };

  const initialValues: IFlagBase = {
    key: flag?.key || '',
    name: flag?.name || '',
    description: flag?.description || '',
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
              navigate(
                `/namespaces/${currentNamespace.key}/flags/${values.key}`
              );
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
                    enabled={enabled}
                    handleChange={(e) => {
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
                    handleChange={(e) => {
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
                    handleChange={(e) => {
                      const formatted = stringAsKey(e.target.value);
                      formik.setFieldValue('key', formatted);
                    }}
                  />
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
                  disabled={
                    !(formik.dirty && formik.isValid && !formik.isSubmitting)
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
