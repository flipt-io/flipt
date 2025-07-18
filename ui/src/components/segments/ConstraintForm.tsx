import * as Dialog from '@radix-ui/react-dialog';
import { addMinutes, format, formatISO, isValid, parseISO } from 'date-fns';
import { Form, Formik, useField, useFormikContext } from 'formik';
import { CircleHelpIcon, XIcon } from 'lucide-react';
import { forwardRef, useContext, useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Link } from 'react-router';
import * as Yup from 'yup';

import { selectTimezone } from '~/app/preferences/preferencesSlice';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import MoreInfo from '~/components/MoreInfo';
import Input from '~/components/forms/Input';
import Select from '~/components/forms/Select';
import Tags from '~/components/forms/Tags';

import {
  ConstraintBooleanOperators,
  ConstraintDateTimeOperators,
  ConstraintEntityIdOperators,
  ConstraintNumberOperators,
  ConstraintStringOperators,
  ConstraintType,
  IConstraint,
  NoValueOperators,
  constraintTypeToLabel
} from '~/types/Constraint';
import { Timezone } from '~/types/Preferences';

import { useError } from '~/data/hooks/error';
import {
  jsonNumberArrayValidation,
  jsonStringArrayValidation,
  requiredValidation
} from '~/data/validations';

import { SegmentFormContext } from './SegmentFormContext';

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
    case ConstraintType.ENTITY_ID:
      opts = ConstraintEntityIdOperators;
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
  placeholder?: string;
};

type ConstraintArrayInputProps = ConstraintInputProps & { type?: string };

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
          className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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

function ConstraintValueArrayInput(props: ConstraintArrayInputProps) {
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
          className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
        >
          Values
        </label>
      </div>
      <div className="sm:col-span-2">
        <Tags {...props} {...field} />
      </div>
    </div>
  );
}

function ConstraintValueDateTimeInput(props: ConstraintInputProps) {
  const { setFieldValue } = useFormikContext();
  const timezone = useSelector(selectTimezone);

  const [field] = useField({
    ...props,
    validate: (value: string) => {
      if (!value) return undefined; // Allow empty values initially
      try {
        const m = parseISO(value);
        return isValid(m) ? undefined : 'Value is not a valid datetime';
      } catch (err) {
        return 'Value is not a valid datetime';
      }
    }
  });

  const [fieldDate, setFieldDate] = useState('');
  const [fieldTime, setFieldTime] = useState('');

  useEffect(() => {
    if (field.value) {
      try {
        let m = parseISO(field.value);
        if (isValid(m)) {
          if (timezone === Timezone.UTC) {
            // if utc timezone, then convert to UTC
            m = addMinutes(m, m.getTimezoneOffset());
          }
          setFieldDate(format(m, 'yyyy-MM-dd'));
          setFieldTime(format(m, 'HH:mm'));
        }
      } catch (err) {
        // If parsing fails, leave fields empty
        console.error('Failed to parse date:', err);
      }
    }
  }, [field.value, timezone]);

  useEffect(() => {
    // if both date and time are set, then combine, parse, and set the value
    if (fieldDate && fieldDate.trim() !== '') {
      try {
        let d = `${fieldDate}T00:00:00`;
        if (fieldTime && fieldTime.trim() !== '') {
          d = `${fieldDate}T${fieldTime}:00`;
        }
        let m = parseISO(d);

        if (isValid(m)) {
          if (timezone === Timezone.UTC) {
            // if utc timezone, then convert to UTC
            m = addMinutes(m, -m.getTimezoneOffset());
          }
          setFieldValue(field.name, formatISO(m));
        }
      } catch (err) {
        console.error('Failed to format date:', err);
      }
    }
  }, [timezone, field.name, fieldDate, fieldTime, setFieldValue]);

  return (
    <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
      <div>
        <label
          htmlFor="value"
          className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
        >
          Value
        </label>
        <span
          className="text-xs text-gray-400 dark:text-gray-500"
          id="value-tz"
        >
          <Link
            to="/settings"
            className="group inline-flex items-center text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
          >
            <CircleHelpIcon
              className="-ml-1 h-4 w-4 text-gray-300 group-hover:text-gray-400 dark:text-gray-600 dark:group-hover:text-gray-500"
              aria-hidden="true"
            />
            <span className="ml-1">
              {timezone === Timezone.UTC ? 'UTC' : 'Local'}
            </span>
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
          value={fieldTime}
          onChange={(e) => {
            setFieldTime(e.target.value);
          }}
        />
      </div>
      <input type="hidden" {...props} {...field} />
    </div>
  );
}

const validationSchema = Yup.object({
  property: requiredValidation
})
  .when({
    is: (c: IConstraint) =>
      c.type === ConstraintType.STRING &&
      ['isoneof', 'isnotoneof'].includes(c.operator),
    then: (schema) => schema.shape({ value: jsonStringArrayValidation })
  })
  .when({
    is: (c: IConstraint) =>
      c.type === ConstraintType.NUMBER &&
      ['isoneof', 'isnotoneof'].includes(c.operator),
    then: (schema) => schema.shape({ value: jsonNumberArrayValidation })
  });

type ConstraintFormProps = {
  setOpen: (open: boolean) => void;
  constraint?: (IConstraint & { index: number }) | null;
  onSuccess: () => void;
};

const ConstraintForm = forwardRef((props: ConstraintFormProps, ref: any) => {
  const { setOpen, constraint, onSuccess } = props;

  const isNew = constraint === null;
  const submitPhrase = isNew ? 'Add' : 'Done';
  const title = isNew ? 'New Constraint' : 'Edit Constraint';

  const { setError, clearError } = useError();

  const [hasValue, setHasValue] = useState(
    !NoValueOperators.includes(constraint?.operator || 'eq')
  );

  const { updateConstraint, createConstraint } = useContext(SegmentFormContext);

  const handleSubmit = async (values: IConstraint) => {
    if (isNew) {
      createConstraint(values);
      return;
    }
    updateConstraint(values);
  };

  const getValuePlaceholder = (type: ConstraintType, operator: string) => {
    if (['isoneof', 'isnotoneof'].includes(operator)) {
      const placeholder = {
        [ConstraintType.STRING]: 'Eg: Any text',
        [ConstraintType.NUMBER]: 'Eg: 200',
        [ConstraintType.ENTITY_ID]: 'Eg: Any text'
      }[type as string];
      return placeholder ?? '';
    }

    const placeholder = {
      [ConstraintType.STRING]: 'Eg: Any text',
      [ConstraintType.NUMBER]: 'Eg: 200'
    }[type as string];
    return placeholder ?? '';
  };

  const getConstraintTypeDataFormat = (type: ConstraintType): string =>
    type == ConstraintType.NUMBER ? 'number' : 'text';

  return (
    <Formik
      validateOnChange
      validate={(values: IConstraint) => {
        !isNew && handleSubmit(values);
      }}
      initialValues={{
        id: constraint?.id || '',
        property: constraint?.property || '',
        type: constraint?.type || ConstraintType.STRING,
        operator: constraint?.operator || 'eq',
        value: constraint?.value || '',
        description: constraint?.description || ''
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
                  <MoreInfo href="https://docs.flipt.io/v2/concepts#constraints">
                    Learn more about constraints
                  </MoreInfo>
                </div>
                <div className="flex h-7 items-center">
                  <button
                    type="button"
                    className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
                    onClick={() => {
                      setOpen(false);
                    }}
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
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
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
                      const previousType = formik.values.type;

                      const type = e.target.value as ConstraintType;
                      formik.setFieldValue('type', type);

                      if (e.target.value === ConstraintType.BOOLEAN) {
                        formik.setFieldValue('operator', 'true');
                        setHasValue(false);
                      } else {
                        formik.setFieldValue('operator', 'eq');
                        setHasValue(true);
                      }

                      if (e.target.value === ConstraintType.ENTITY_ID) {
                        formik.setFieldValue('property', 'entityId');
                      } else if (
                        previousType === ConstraintType.ENTITY_ID &&
                        e.target.value !== ConstraintType.ENTITY_ID
                      ) {
                        formik.setFieldValue('property', '');
                      }
                    }}
                  />
                </div>
              </div>
              {formik.values.type !== ConstraintType.ENTITY_ID && (
                <>
                  <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                    <div>
                      <label
                        htmlFor="property"
                        className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                      >
                        Property
                      </label>
                    </div>
                    <div className="sm:col-span-2">
                      <Input name="property" id="property" forwardRef={ref} />
                    </div>
                  </div>
                </>
              )}
              <div className="space-y-1 px-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:space-y-0 sm:px-6 sm:py-5">
                <div>
                  <label
                    htmlFor="operator"
                    className="block text-sm font-medium text-gray-900 dark:text-gray-100 sm:mt-px sm:pt-2"
                  >
                    Operator
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <ConstraintOperatorSelect
                    id="operator"
                    name="operator"
                    type={formik.values.type}
                    onChange={(e) => {
                      const noValue = NoValueOperators.includes(e.target.value);
                      setHasValue(!noValue);
                      if (noValue) {
                        formik.setFieldValue('value', '');
                      }
                    }}
                  />
                </div>
              </div>
              {hasValue &&
                (formik.values.type === ConstraintType.DATETIME ? (
                  <ConstraintValueDateTimeInput name="value" id="value" />
                ) : formik.values.operator == 'isoneof' ||
                  formik.values.operator == 'isnotoneof' ? (
                  <ConstraintValueArrayInput
                    name="value"
                    id="value"
                    type={getConstraintTypeDataFormat(formik.values.type)}
                    placeholder={getValuePlaceholder(
                      formik.values.type,
                      formik.values.operator
                    )}
                  />
                ) : (
                  <ConstraintValueInput
                    name="value"
                    id="value"
                    placeholder={getValuePlaceholder(
                      formik.values.type,
                      formik.values.operator
                    )}
                  />
                ))}
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
              <Button
                variant="secondary"
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
