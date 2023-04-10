import { Dialog } from '@headlessui/react';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik, useField, useFormikContext } from 'formik';
import { forwardRef, useEffect, useState } from 'react';
import * as Yup from 'yup';
import Button from '~/components/forms/Button';
import Input from '~/components/forms/Input';
import Select from '~/components/forms/Select';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { createConstraint, updateConstraint } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { requiredValidation } from '~/data/validations';
import {
  ComparisonType,
  ConstraintBooleanOperators,
  ConstraintNumberOperators,
  ConstraintStringOperators,
  IConstraint,
  IConstraintBase,
  NoValueOperators
} from '~/types/Constraint';

const constraintComparisonTypes = () =>
  (Object.keys(ComparisonType) as Array<keyof typeof ComparisonType>).map(
    (t) => ({
      value: t,
      label: ComparisonType[t]
    })
  );

const constraintOperators = (c: string) => {
  let opts: Record<string, string> = {};
  switch (ComparisonType[c as keyof typeof ComparisonType]) {
    case ComparisonType.STRING_COMPARISON_TYPE:
      opts = ConstraintStringOperators;
      break;
    case ComparisonType.NUMBER_COMPARISON_TYPE:
      opts = ConstraintNumberOperators;
      break;
    case ComparisonType.BOOLEAN_COMPARISON_TYPE:
      opts = ConstraintBooleanOperators;
      break;
  }
  return Object.entries(opts).map(([k, v]) => ({
    value: k,
    label: v
  }));
};

type InputProps = {
  name: string;
  id: string;
};

function ConstraintOperatorSelect(props: InputProps) {
  const {
    values: { type },
    setFieldValue
  } = useFormikContext<{ type: string }>();

  const [field] = useField(props);

  return (
    <Select
      className="mt-1"
      {...field}
      {...props}
      handleChange={(e) => {
        setFieldValue(field.name, e.target.value);
      }}
      options={constraintOperators(type)}
    />
  );
}

function ConstraintValueField(props: InputProps) {
  const [show, setShow] = useState(true);
  const {
    values: { type, operator },
    dirty
  } = useFormikContext<{ type: string; operator: string }>();

  const [field] = useField({
    ...props,
    validate: (value) => {
      if (!show) {
        return undefined;
      }

      // value is required only if shown
      return value ? undefined : 'Value is required';
    }
  });

  // show/hide value field based on operator
  useEffect(() => {
    if (type === 'BOOLEAN_COMPARISON_TYPE') {
      setShow(false);
      return;
    }
    const noValue = NoValueOperators.includes(operator);
    setShow(!noValue);
  }, [type, operator, field.name, dirty]);

  if (!show) {
    return <></>;
  }

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="value"
          className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
        >
          Value
        </label>
      </div>
      <div className="sm:col-span-2">
        <Input {...props} {...field} />
      </div>
    </div>
  );
}

type ConstraintFormProps = {
  setOpen: (open: boolean) => void;
  segmentKey: string;
  constraint?: IConstraint;
  onSuccess: () => void;
};

const ConstraintForm = forwardRef((props: ConstraintFormProps, ref: any) => {
  const { setOpen, segmentKey, constraint, onSuccess } = props;

  const isNew = constraint === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';
  const title = isNew ? 'New Constraint' : 'Edit Constraint';

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const initialValues: IConstraintBase = {
    property: constraint?.property || '',
    type: constraint?.type || ('STRING_COMPARISON_TYPE' as ComparisonType),
    operator: constraint?.operator || 'eq',
    value: constraint?.value || ''
  };

  const handleSubmit = async (values: IConstraintBase) => {
    if (isNew) {
      return createConstraint(currentNamespace.key, segmentKey, values);
    }
    return updateConstraint(
      currentNamespace.key,
      segmentKey,
      constraint?.id,
      values
    );
  };

  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize={true}
      onSubmit={(values, { setSubmitting }) => {
        handleSubmit(values)
          .then(() => {
            clearError();
            setSuccess(
              `Successfully ${submitPhrase.toLocaleLowerCase()}d constraint`
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
        property: requiredValidation
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
                  <MoreInfo href="https://www.flipt.io/docs/concepts#constraints">
                    Learn more about constraints
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
                    htmlFor="property"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Property
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Input name="property" id="property" forwardRef={ref} />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Type
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <Select
                    name="type"
                    id="type"
                    className="mt-1"
                    value={formik.values.type}
                    options={constraintComparisonTypes()}
                    handleChange={(e) => {
                      formik.setFieldValue('type', e.target.value);

                      if (e.target.value === 'BOOLEAN_COMPARISON_TYPE') {
                        formik.setFieldValue('operator', 'true');
                      } else {
                        formik.setFieldValue('operator', 'eq');
                      }
                    }}
                  />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="operator"
                    className="block text-sm font-medium text-gray-900 sm:mt-px sm:pt-2"
                  >
                    Operator
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <ConstraintOperatorSelect id="operator" name="operator" />
                </div>
              </div>
              <ConstraintValueField name="value" id="value" />
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
                {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
              </Button>
            </div>
          </div>
        </Form>
      )}
    </Formik>
  );
});

ConstraintForm.displayName = 'ConstraintForm';
export default ConstraintForm;
