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
    description: 'All constraints must match'
  },
  {
    id: SegmentMatchType.ANY,
    name: 'Any',
    description: 'At least one constraints must match'
  }
];

const segmentValidationSchema = Yup.object({
  key: keyValidation,
  name: requiredValidation
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
        environmentKey: environment.name,
        namespaceKey: namespace.key,
        values,
        revision
      }).unwrap();
    }
    return updateSegment({
      environmentKey: environment.name,
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
        const { constraints } = formik.values;
        const disableSave = !(
          formik.dirty &&
          formik.isValid &&
          !formik.isSubmitting
        );

        const form = (
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
                                formik.setFieldValue('matchType', matchType.id);
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

              {segment && (
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
                <Button
                  variant="primary"
                  className="ml-3 min-w-[80px]"
                  type="submit"
                  disabled={disableSave}
                >
                  {formik.isSubmitting ? <Loading isPrimary /> : submitPhrase}
                </Button>
              </div>
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
