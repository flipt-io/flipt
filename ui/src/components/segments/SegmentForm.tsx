import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/react/20/solid';
import { Form, Formik } from 'formik';
import { useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router';
import * as Yup from 'yup';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import {
  useCreateSegmentMutation,
  useUpdateSegmentMutation
} from '~/app/segments/segmentsApi';

import { Button } from '~/components/Button';
import Loading from '~/components/Loading';
import { UnsavedChangesModalWrapper } from '~/components/UnsavedChangesModal';
import Constraints from '~/components/constraints/Constraints';
import Input from '~/components/forms/Input';

import { ISegment, SegmentMatchType } from '~/types/Segment';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';
import { keyValidation, requiredValidation } from '~/data/validations';
import {
  cls,
  copyTextToClipboard,
  getRevision,
  stringAsKey
} from '~/utils/helpers';

import { SegmentFormProvider } from './SegmentFormContext';

const segmentMatchTypes = [
  {
    id: SegmentMatchType.ALL,
    name: 'All',
    description:
      'All constraints must match for the segment to be considered a match'
  },
  {
    id: SegmentMatchType.ANY,
    name: 'Any',
    description:
      'At least one constraint must match for the segment to be considered a match'
  }
];

function SegmentTypeSelector({
  selectedType,
  onTypeSelect
}: {
  selectedType: SegmentMatchType | '';
  onTypeSelect: (type: SegmentMatchType) => void;
}) {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-medium text-gray-900">Match Type</h2>
        <p className="mt-1 text-sm text-gray-500">
          Select how constraints should be evaluated for this segment
        </p>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {segmentMatchTypes.map((matchType) => (
          <div
            key={matchType.id}
            onClick={() => onTypeSelect(matchType.id)}
            className={cls(
              'relative flex cursor-pointer flex-col rounded-lg border p-4 shadow-sm focus:outline-none hover:border-violet-500',
              {
                'border-violet-500 ring ring-violet-500':
                  selectedType === matchType.id,
                'border-gray-300': selectedType !== matchType.id
              }
            )}
          >
            <div className="flex flex-1">
              <div className="flex flex-col">
                <div className="flex items-center">
                  <span className="text-sm font-medium text-gray-900">
                    {matchType.name}
                  </span>
                </div>
                <p className="mt-2 flex items-center text-sm text-gray-500">
                  {matchType.description}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const segmentValidationSchema = Yup.object({
  key: keyValidation,
  name: requiredValidation,
  matchType: Yup.string().required('Please select a match type')
});

type SegmentFormProps = {
  segment?: ISegment;
};

export default function SegmentForm(props: SegmentFormProps) {
  const { segment } = props;

  const isNew = segment === undefined;
  const submitPhrase = isNew ? 'Create' : 'Update';

  const navigate = useNavigate();
  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const revision = getRevision();

  const [createSegment] = useCreateSegmentMutation();
  const [updateSegment] = useUpdateSegmentMutation();

  const handleSubmit = (values: ISegment) => {
    if (isNew) {
      return createSegment({
        environmentKey: environment.key,
        namespaceKey: namespace.key,
        values,
        revision
      }).unwrap();
    }
    return updateSegment({
      environmentKey: environment.key,
      namespaceKey: namespace.key,
      segmentKey: segment?.key,
      values,
      revision
    }).unwrap();
  };

  const initialValues: ISegment = {
    key: segment?.key || '',
    name: segment?.name || '',
    description: segment?.description || '',
    matchType: segment?.matchType || SegmentMatchType.ALL,
    constraints: segment?.constraints || []
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
              `Successfully ${submitPhrase.toLocaleLowerCase()}d segment`
            );

            if (isNew) {
              navigate(`/namespaces/${namespace.key}/segments/${values.key}`);
              return;
            }
          })
          .catch((err) => {
            setError(err);
          })
          .finally(() => {
            setSubmitting(false);
          });
      }}
      validationSchema={segmentValidationSchema}
    >
      {(formik) => {
        const { constraints, matchType } = formik.values;
        const disableSave = !(
          formik.dirty &&
          formik.isValid &&
          !formik.isSubmitting &&
          (isNew
            ? matchType === SegmentMatchType.ALL ||
              matchType === SegmentMatchType.ANY
            : true)
        );

        const form = (
          <Form className="space-y-6 p-1 sm:overflow-hidden sm:rounded-md">
            <SegmentTypeSelector
              selectedType={formik.values.matchType}
              onTypeSelect={(type) => {
                formik.setFieldValue('matchType', type);
              }}
            />

            {(!isNew || formik.values.matchType) && (
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
                    onChange={(e) => {
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
                  <div
                    className={cls({
                      'flex items-center justify-between': !isNew
                    })}
                  >
                    <Input
                      className={cls('mt-1', { 'md:mr-2': !isNew })}
                      name="key"
                      id="key"
                      disabled={!isNew}
                      onChange={(e) => {
                        const formatted = stringAsKey(e.target.value);
                        formik.setFieldValue('key', formatted);
                      }}
                    />
                    {!isNew && (
                      <button
                        aria-label="Copy to clipboard"
                        title="Copy to Clipboard"
                        className="hidden md:block"
                        onClick={(e) => {
                          e.preventDefault();
                          copyTextToClipboard(segment?.key || '');
                          setKeyCopied(true);
                          setTimeout(() => {
                            setKeyCopied(false);
                          }, 2000);
                        }}
                      >
                        <CheckIcon
                          className={cls(
                            'absolute m-auto h-5 w-5 justify-center align-middle text-green-400 transition-opacity duration-300 ease-in-out',
                            {
                              'visible opacity-100': keyCopied,
                              'invisible opacity-0': !keyCopied
                            }
                          )}
                        />
                        <ClipboardDocumentIcon
                          className={cls(
                            'm-auto h-5 w-5 justify-center align-middle text-gray-300 transition-opacity duration-300 ease-in-out hover:text-gray-400',
                            {
                              'visible opacity-100': !keyCopied,
                              'invisible opacity-0': keyCopied
                            }
                          )}
                        />
                      </button>
                    )}
                  </div>
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
            )}

            {segment && matchType && (
              <>
                <div className="mt-3 flex flex-row sm:mt-5">
                  <Constraints constraints={constraints!} />
                </div>
              </>
            )}

            <div className="flex justify-end">
              <Button type="button" onClick={() => navigate(-1)}>
                Cancel
              </Button>
              {(!isNew || formik.values.matchType) && (
                <div className="relative inline-block">
                  <Button
                    variant="primary"
                    className="ml-3 min-w-[80px]"
                    type="submit"
                    disabled={disableSave}
                  >
                    {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                  </Button>
                  {formik.dirty && formik.isValid && (
                    <div className="absolute -right-1 -top-1 h-3 w-3">
                      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-violet-100 opacity-75"></span>
                    </div>
                  )}
                </div>
              )}
            </div>
          </Form>
        );
        return (
          <SegmentFormProvider formik={formik}>
            {isNew ? (
              form
            ) : (
              <UnsavedChangesModalWrapper formik={formik}>
                {form}
              </UnsavedChangesModalWrapper>
            )}
          </SegmentFormProvider>
        );
      }}
    </Formik>
  );
}
