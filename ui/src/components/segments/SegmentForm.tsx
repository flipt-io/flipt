import { Form, Formik } from 'formik';
import { useNavigate } from 'react-router-dom';
import * as Yup from 'yup';
import Button from '~/components/forms/Button';
import Input from '~/components/forms/Input';
import Loading from '~/components/Loading';
import { createSegment, updateSegment } from '~/data/api';
import { useError } from '~/data/hooks/error';
import useNamespace from '~/data/hooks/namespace';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import { ISegment, ISegmentBase, SegmentMatchType } from '~/types/Segment';
import { stringAsKey } from '~/utils/helpers';

const segmentMatchTypes = [
  {
    id: 'ALL_MATCH_TYPE',
    name: SegmentMatchType.ALL_MATCH_TYPE,
    description: 'All constraints must match'
  },
  {
    id: 'ANY_MATCH_TYPE',
    name: SegmentMatchType.ANY_MATCH_TYPE,
    description: 'At least one constraints must match'
  }
];

type SegmentFormProps = {
  segment?: ISegment;
  segmentChanged?: () => void;
};

export default function SegmentForm(props: SegmentFormProps) {
  const { segment, segmentChanged } = props;

  const isNew = segment === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';

  const navigate = useNavigate();
  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const { currentNamespace } = useNamespace();

  const handleSubmit = (values: ISegmentBase) => {
    if (isNew) {
      return createSegment(currentNamespace.key, values);
    }
    return updateSegment(currentNamespace.key, segment?.key, values);
  };

  const initialValues: ISegmentBase = {
    key: segment?.key || '',
    name: segment?.name || '',
    description: segment?.description || '',
    matchType: segment?.matchType || ('ALL_MATCH_TYPE' as SegmentMatchType)
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
              `Successfully ${submitPhrase.toLocaleLowerCase()}d segment`
            );

            if (isNew) {
              navigate(
                `/namespaces/${currentNamespace.key}/segments/${values.key}`
              );
              return;
            }
            segmentChanged && segmentChanged();
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
      {(formik) => (
        <Form className="px-1 sm:overflow-hidden sm:rounded-md">
          <div className="space-y-6">
            <div className="grid grid-cols-3 gap-6">
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
                      formik.setFieldValue('key', stringAsKey(e.target.value));
                    }
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
                <label
                  htmlFor="matchType"
                  className="block text-sm font-medium text-gray-700"
                >
                  Match Type
                </label>
                <fieldset className="mt-2">
                  <legend className="sr-only">Match Type</legend>
                  <div className="space-y-5">
                    {segmentMatchTypes.map((matchType) => (
                      <div
                        key={matchType.id}
                        className="relative flex items-start"
                      >
                        <div className="flex h-5 items-center">
                          <input
                            id={matchType.id}
                            aria-describedby={`${matchType.id}-description`}
                            name="matchType"
                            type="radio"
                            className="h-4 w-4 border-gray-300 text-violet-400 focus:ring-violet-400"
                            onChange={() => {
                              formik.setFieldValue(
                                'matchType',
                                matchType.id as SegmentMatchType
                              );
                            }}
                            checked={matchType.id === formik.values.matchType}
                            value={matchType.id}
                          />
                        </div>
                        <div className="ml-3 text-sm">
                          <label
                            htmlFor={matchType.id}
                            className="font-medium text-gray-700"
                          >
                            {matchType.name}
                          </label>
                          <p
                            id={`${matchType.id}-description`}
                            className="text-gray-500"
                          >
                            {matchType.description}
                          </p>
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
}
