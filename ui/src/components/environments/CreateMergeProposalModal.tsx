import { Field, Form, Formik } from 'formik';
import * as Yup from 'yup';

import {
  useListBranchEnvironmentChangesQuery,
  useProposeEnvironmentMutation
} from '~/app/environments/environmentsApi';

import Loading from '~/components/Loading';
import Modal from '~/components/Modal';
import { Button } from '~/components/ui/button';

import { IEnvironment } from '~/types/Environment';

import { useError } from '~/data/hooks/error';
import { useSuccess } from '~/data/hooks/success';

interface CreateMergeProposalModalProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  environment: IEnvironment;
}

const validationSchema = Yup.object().shape({
  description: Yup.string()
    .optional()
    .max(500, 'Description must be at most 500 characters')
    .trim(),
  draft: Yup.boolean()
});

const MAX_COMMITS = 10;

export function CreateMergeProposalModal({
  open,
  setOpen,
  environment
}: CreateMergeProposalModalProps) {
  const { data, isLoading, isError } = useListBranchEnvironmentChangesQuery(
    {
      baseEnvironmentKey: environment.configuration?.base ?? '',
      environmentKey: environment.key
    },
    { skip: !open }
  );

  const { setError, clearError } = useError();
  const { setSuccess } = useSuccess();

  const [proposeEnvironment] = useProposeEnvironmentMutation();

  const handleProposeEnvironment = async (values: {
    description: string;
    draft: boolean;
  }) => {
    try {
      await proposeEnvironment({
        baseEnvironmentKey: environment.configuration?.base ?? '',
        environmentKey: environment.key,
        body: values.description,
        draft: values.draft
      }).unwrap();
      setOpen(false);
      clearError();
      setSuccess('Merge proposal created successfully');
    } catch (error) {
      setError(error);
    }
  };

  return (
    <Modal open={open} setOpen={setOpen}>
      <div className="p-6">
        <h2 className="text-2xl font-bold mb-1">Create Merge Proposal</h2>
        <p className="text-muted-foreground mb-6">
          Propose merging changes from this environment back into{' '}
          <span className="font-semibold">
            {environment.configuration?.base}
          </span>
          .
        </p>
        <div className="bg-muted rounded-md px-3 py-2 mb-6 border border-muted-foreground/10 min-h-[60px]">
          {isLoading && (
            <div className="text-sm text-muted-foreground">
              Loading changes…
            </div>
          )}
          {isError && (
            <div className="text-sm text-destructive">
              Failed to load changes.
            </div>
          )}
          {!isLoading && !isError && data && (
            <div className="max-h-64 overflow-y-auto">
              <ul className="divide-y divide-muted-foreground/10">
                {data.changes.length === 0 && (
                  <li className="py-2 text-sm text-muted-foreground">
                    No changes to merge.
                  </li>
                )}
                {data.changes.map((change) => (
                  <li key={change.revision} className="flex items-center py-1">
                    <span className="font-mono text-xs text-muted-foreground bg-gray-100 rounded px-2 py-0.5 mr-3">
                      {change.scmUrl ? (
                        <a
                          href={change.scmUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-brand"
                        >
                          {change.revision.slice(0, 7)}
                        </a>
                      ) : (
                        change.revision.slice(0, 7)
                      )}
                    </span>
                    <span>
                      <span className="text-sm font-normal">
                        {change.message}
                      </span>
                      {change.authorName && (
                        <span className="ml-2 text-xs text-muted-foreground">
                          by {change.authorName}
                        </span>
                      )}
                    </span>
                  </li>
                ))}
                {data.changes.length === MAX_COMMITS && (
                  <li className="py-2 text-center text-xs text-muted-foreground">
                    ...and more
                  </li>
                )}
              </ul>
            </div>
          )}
        </div>
        <Formik
          initialValues={{ description: '', draft: false }}
          validationSchema={validationSchema}
          onSubmit={async (values, actions) => {
            const trimmed = values.description.trim().slice(0, 500);
            const submitValues = { ...values, description: trimmed };
            await handleProposeEnvironment(submitValues);
            actions.setSubmitting(false);
          }}
        >
          {(formik) => {
            const disableSave = !(formik.isValid && !formik.isSubmitting);
            return (
              <Form>
                <div className="mb-6">
                  <label
                    className="block text-sm font-medium mb-1"
                    htmlFor="proposal-desc"
                  >
                    Proposal Description{' '}
                    <span className="text-muted-foreground">(optional)</span>
                  </label>
                  <Field
                    as="textarea"
                    id="proposal-desc"
                    name="description"
                    className="w-full rounded-md border px-3 py-2 text-sm focus:ring-2 focus:ring-brand focus:border-brand transition"
                    rows={3}
                    placeholder="Add context or reasoning for this merge proposal..."
                    maxLength={500}
                    onChange={formik.handleChange}
                    value={formik.values.description}
                  />
                  <div className="text-xs text-muted-foreground text-right mt-1">
                    {formik.values.description.trim().length}/500
                  </div>
                  {formik.errors.description && formik.touched.description && (
                    <div className="text-xs text-destructive mt-1">
                      {formik.errors.description}
                    </div>
                  )}
                </div>
                <div className="mb-4 flex items-center">
                  <Field
                    type="checkbox"
                    id="draft-proposal"
                    name="draft"
                    className="mr-2 accent-brand"
                  />
                  <label
                    htmlFor="draft-proposal"
                    className="text-sm select-none cursor-pointer"
                  >
                    Open as <span className="font-medium">Draft</span>
                  </label>
                </div>
                <div className="flex justify-end gap-2">
                  <Button
                    variant="outline"
                    type="button"
                    onClick={() => setOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="primary"
                    type="submit"
                    disabled={disableSave}
                  >
                    {formik.isSubmitting ? (
                      <Loading isPrimary />
                    ) : (
                      'Submit Proposal'
                    )}
                  </Button>
                </div>
              </Form>
            );
          }}
        </Formik>
      </div>
    </Modal>
  );
}
