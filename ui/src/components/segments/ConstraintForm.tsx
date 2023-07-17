import { Dialog } from '@headlessui/react';
import { QuestionMarkCircleIcon } from '@heroicons/react/20/solid';
import { XMarkIcon } from '@heroicons/react/24/outline';
import { Form, Formik, useField, useFormikContext } from 'formik';
import moment from 'moment';
import { forwardRef, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router-dom';
import * as Yup from 'yup';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { selectTimezone } from '~/app/preferences/preferencesSlice';
import Button from '~/components/forms/buttons/Button';
import Input from '~/components/forms/Input';
import Select from '~/components/forms/Select';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import { createConstraint, updateConstraint } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { requiredValidation } from '~/data/validations';
import {
  ConstraintBooleanOperators,
  ConstraintDateTimeOperators,
  ConstraintNumberOperators,
  ConstraintStringOperators,
  ConstraintType,
  constraintTypeToLabel,
  IConstraint,
  IConstraintBase,
  NoValueOperators
} from '~/types/Constraint';
import { Timezone } from '~/types/Preferences';

const constraintComparisonTypes = () =>
  (Object.keys(ConstraintType) as Array<keyof typeof ConstraintType>).map(
    (t) => ({
      value: ConstraintType[t],
      label: constraintTypeToLabel(ConstraintType[t])
    })
  );

const constraintOperators = (c: string) => {
  let opts: Record<string, string> = {};
  switch (c as ConstraintType) {
    case ConstraintType.STRING:
      opts = ConstraintStringOperators;
      break;
    case ConstraintType.NUMBER:
      opts = ConstraintNumberOperators;
      break;
    case ConstraintType.BOOLEAN:
      opts = ConstraintBooleanOperators;
      break;
    case ConstraintType.DATETIME:
      opts = ConstraintDateTimeOperators;
      break;
  }
  return Object.entries(opts).map(([k, v]) => ({
    value: k,
    label: v
  }));
};

type ConstraintInputProps = {
  name: string;
  id: string;
};

type ConstraintOperatorSelectProps = ConstraintInputProps & {
  onChange: (e: React.ChangeEvent<any>) => void;
  type: string;
};

function ConstraintOperatorSelect(props: ConstraintOperatorSelectProps) {
  const { onChange, type } = props;

  const { setFieldValue } = useFormikContext();

  const [field] = useField(props);

  return (
    <Select
      className="mt-1"
      {...field}
      {...props}
      onChange={(e) => {
        setFieldValue(field.name, e.target.value);
        onChange(e);
      }}
      options={constraintOperators(type)}
    />
  );
}

function ConstraintValueInput(props: ConstraintInputProps) {
  const [field] = useField({
    ...props,
    validate: (value) => {
      // value is required only if shown
      return value ? undefined : 'Value is required';
    }
  });

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="value"
          className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
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

function ConstraintValueDateTimeInput(props: ConstraintInputProps) {
  const { setFieldValue } = useFormikContext();
  const timezone = useSelector(selectTimezone);

  const [field] = useField({
    ...props,
    validate: (value) => {
      const m = moment(value);
      return m.isValid() ? undefined : 'Value is not a valid datetime';
    }
  });

  const [fieldDate, setFieldDate] = useState(field.value?.split('T')[0] || '');
  const [fieldTime, setFieldTime] = useState(field.value?.split('T')[1] || '');

  useEffect(() => {
    // if both date and time are set, then combine, parse, and set the value
    if (
      fieldDate &&
      fieldDate.trim() !== '' &&
      fieldTime &&
      fieldTime.trim() !== ''
    ) {
      if (timezone === Timezone.LOCAL) {
        // if local timezone, then parse as local (moment default) and convert to UTC
        const m = moment(`${fieldDate}T${fieldTime}`);
        setFieldValue(field.name, m.utc().format());
        return;
      }

      // otherwise, parse as UTC
      const m = moment.utc(`${fieldDate}T${fieldTime}`);
      setFieldValue(field.name, m.format());
      return;
    }

    // otherwise, if only date is set, then parse and set the value
    if (fieldDate && fieldDate.trim() !== '') {
      const m = moment(fieldDate);
      setFieldValue(field.name, m.utc().format());
    }
  }, [timezone, field.name, fieldDate, fieldTime, setFieldValue]);

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="value"
          className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
        >
          Value
        </label>
        <span className="text-gray-400 text-xs" id="value-tz">
          <Link
            to="/settings"
            className="group text-gray-400 inline-flex items-center hover:text-gray-500"
          >
            <QuestionMarkCircleIcon
              className="text-gray-300 -ml-1 h-4 w-4 group-hover:text-gray-400"
              aria-hidden="true"
            />
            <span className="ml-1">{timezone}</span>
          </Link>
        </span>
      </div>
      <div className="sm:col-span-1">
        <Input
          type="date"
          id="valueDate"
          name="valueDate"
          value={fieldDate}
          onChange={(e) => {
            setFieldDate(e.target.value);
          }}
        />
      </div>
      <div className="sm:col-span-1">
        <Input
          type="time"
          id="valueTime"
          name="valueTime"
          value={
            timezone === Timezone.LOCAL
              ? moment(fieldTime, 'HH:mm Z').format('HH:mm')
              : moment.utc(fieldTime, 'HH:mm Z').format('HH:mm')
          }
          onChange={(e) => {
            setFieldTime(e.target.value);
          }}
        />
      </div>
      <input type="hidden" {...props} {...field} />
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

  const [hasValue, setHasValue] = useState(true);
  const [type, setType] = useState(constraint?.type || ConstraintType.STRING);

  const namespace = useSelector(selectCurrentNamespace);

  const initialValues = {
    property: constraint?.property || '',
    type: constraint?.type || ConstraintType.STRING,
    operator: constraint?.operator || 'eq',
    value: constraint?.value || '',
    description: constraint?.description || ''
  };

  const handleSubmit = (values: IConstraintBase) => {
    if (isNew) {
      return createConstraint(namespace.key, segmentKey, values);
    }
    return updateConstraint(namespace.key, segmentKey, constraint?.id, values);
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
        <Form className="bg-white flex h-full flex-col overflow-y-scroll shadow-xl">
          <div className="flex-1">
            <div className="bg-gray-50 px-4 py-6 sm:px-6">
              <div className="flex items-start justify-between space-x-3">
                <div className="space-y-1">
                  <Dialog.Title className="text-gray-900 text-lg font-medium">
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
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
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
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
                    onChange={(e) => {
                      const type = e.target.value as ConstraintType;
                      formik.setFieldValue('type', type);
                      setType(type);

                      if (e.target.value === ConstraintType.BOOLEAN) {
                        formik.setFieldValue('operator', 'true');
                        setHasValue(false);
                      } else {
                        formik.setFieldValue('operator', 'eq');
                        setHasValue(true);
                      }
                    }}
                  />
                </div>
              </div>
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="operator"
                    className="text-gray-900 block text-sm font-medium sm:mt-px sm:pt-2"
                  >
                    Operator
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <ConstraintOperatorSelect
                    id="operator"
                    name="operator"
                    type={type}
                    onChange={(e) => {
                      const noValue = NoValueOperators.includes(e.target.value);
                      setHasValue(!noValue);
                    }}
                  />
                </div>
              </div>
              {hasValue &&
                (type === ConstraintType.DATETIME ? (
                  <ConstraintValueDateTimeInput name="value" id="value" />
                ) : (
                  <ConstraintValueInput name="value" id="value" />
                ))}
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
