import { Field, Form, Formik } from 'formik';
import * as Yup from 'yup';

import {
  useListBranchEnvironmentChangesQuery,
  useProposeEnvironmentMutation
} from '~/app/environments/environmentsApi';

import { Badge } from '~/components/Badge';
import { Button } from '~/components/Button';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '~/components/Dialog';
import Loading from '~/components/Loading';

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
      environmentKey: environment.configuration?.base ?? '',
      key: environment.key
    },
    { skip: !open, refetchOnMountOrArgChange: true }
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
        environmentKey: environment.configuration?.base ?? '',
        key: environment.key,
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
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Merge Proposal</DialogTitle>
          <DialogDescription>
            Propose merging changes from this environment back into{' '}
            <span className="font-semibold">
              {environment.configuration?.base}
            </span>
            .
          </DialogDescription>
        </DialogHeader>
        {isError ? (
          <div className="rounded-md px-3 py-2 border border-destructive/50 min-h-[60px] bg-destructive/5 flex flex-col items-center justify-center text-center">
            <span className="text-base font-semibold text-destructive mb-1">
              Failed to load changes
            </span>
            <span className="text-sm text-destructive mb-2">
              You will not be able to create a merge proposal until changes can
              be loaded.
            </span>
          </div>
        ) : (
          <div className="bg-muted/80 dark:bg-input/10 rounded-md p-2 border border-muted-foreground/10 min-h-[60px]">
            {isLoading && (
              <div className="text-sm text-muted-foreground">
                Loading changes…
              </div>
            )}
            {!isLoading && data && (
              <div className="max-h-64 overflow-y-auto">
                <ul className="divide-y divide-muted-foreground/10">
                  {data.changes.length === 0 && (
                    <li className="py-2 text-sm text-muted-foreground">
                      No changes to merge.
                    </li>
                  )}
                  {data.changes.map((change) => (
                    <li
                      key={change.revision}
                      className="flex items-center py-2 pl-1 gap-2"
                    >
                      <Badge variant="outlinemuted">
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
                      </Badge>
                      <span className="text-sm font-normal">
                        {change.message}
                      </span>
                      {change.authorName && (
                        <span className="text-xs text-muted-foreground">
                          by {change.authorName}
                        </span>
                      )}
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
        )}
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
            const disableSave =
              !(formik.isValid && !formik.isSubmitting) ||
              isError ||
              data?.changes?.length == 0;
            return (
              <Form>
                <div>
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
                    className="w-full rounded-md border-input bg-secondary/20 dark:bg-input/20 px-3 py-2 text-sm focus:ring-2 focus:ring-brand focus:border-brand transition disabled:opacity-80 disabled:cursor-not-allowed"
                    rows={3}
                    placeholder="Add context or reasoning for this merge proposal..."
                    maxLength={500}
                    onChange={formik.handleChange}
                    value={formik.values.description}
                    disabled={isError}
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
                    className="border mr-2 rounded checked:bg-brand accent-brand focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[1px] focus:ring-2 focus:ring-brand focus:border-brand transition disabled:opacity-80 disabled:cursor-not-allowed"
                    disabled={isError}
                  />
                  <label
                    htmlFor="draft-proposal"
                    className="text-sm select-none cursor-pointer aria-disabled:opacity-80 aria-disabled:cursor-not-allowed"
                    aria-disabled={isError}
                  >
                    Open as <span className="font-medium">Draft</span>
                  </label>
                </div>
                <DialogFooter>
                  <DialogClose>Cancel</DialogClose>
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
                </DialogFooter>
              </Form>
            );
          }}
        </Formik>
      </DialogContent>
    </Dialog>
  );
}
