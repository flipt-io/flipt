import { Form, Formik } from 'formik';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  useDeleteBranchEnvironmentMutation
} from '~/app/environments/environmentsApi';

import { BaseInput } from '~/components/BaseInput';
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

import { IEnvironment } from '~/types/Environment';

import { useError } from '~/data/hooks/error';
import { useAppDispatch } from '~/data/hooks/store';
import { useSuccess } from '~/data/hooks/success';

interface DeleteBranchModalProps {
  open: boolean;
  setOpen: (open: boolean) => void;
  environment: IEnvironment;
}

export default function DeleteBranchModal({
  open,
  setOpen,
  environment
}: DeleteBranchModalProps) {
  const baseBranch = environment.configuration?.base ?? '';

  const { setSuccess } = useSuccess();
  const { setError, clearError } = useError();

  const dispatch = useAppDispatch();
  const [deleteBranchEnvironment] = useDeleteBranchEnvironmentMutation();

  const validationSchema = Yup.object().shape({
    confirmName: Yup.string()
      .oneOf([environment.key], 'Name does not match')
      .required('Please type the branch environment name to confirm')
  });

  const handleDeleteBranch = async () => {
    try {
      await deleteBranchEnvironment({
        environmentKey: baseBranch,
        key: environment.key
      });
      clearError();
      setSuccess('Branch deleted successfully');
      dispatch(currentEnvironmentChanged(baseBranch));
    } catch (e) {
      setError(e);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <Formik
          initialValues={{ confirmName: '' }}
          validationSchema={validationSchema}
          onSubmit={async (_, actions) => {
            await handleDeleteBranch();
            setOpen(false);
            actions.setSubmitting(false);
          }}
        >
          {(formik) => {
            const { isSubmitting, isValid, errors, touched } = formik;
            return (
              <Form>
                <DialogHeader>
                  <DialogTitle>Delete Branch</DialogTitle>
                  <DialogDescription>
                    This action is{' '}
                    <span className="underline">destructive</span> and cannot be
                    undone.
                  </DialogDescription>
                </DialogHeader>
                <div className="my-4">
                  <p>
                    Deleting the branch will permanently remove all of its data.
                    It will also delete the branch from the remote repository if
                    it exists.
                  </p>
                  <p className="pb-2 pt-4">
                    To confirm deletion, type{' '}
                    <span className="font-bold">{environment.key}</span> in the
                    field below:
                  </p>
                  <BaseInput
                    name="confirmName"
                    type="text"
                    autoCorrect="off"
                    autoComplete="off"
                    placeholder={environment.key}
                    className="w-full rounded-md border-destructive px-3 py-2 text-sm focus-visible:border-destructive/20 focus-visible::ring-2 focus-visible:ring-destructive focus-visible::border-destructive transition mb-2"
                    disabled={isSubmitting}
                    onChange={formik.handleChange}
                    onBlur={formik.handleBlur}
                    value={formik.values.confirmName}
                  />
                  {errors.confirmName &&
                    touched.confirmName &&
                    formik.dirty && (
                      <div className="text-xs text-destructive mb-2">
                        {errors.confirmName}
                      </div>
                    )}
                </div>
                <DialogFooter>
                  <DialogClose disabled={isSubmitting}>Cancel</DialogClose>
                  <Button
                    variant="destructive"
                    type="submit"
                    disabled={!isValid || isSubmitting}
                  >
                    {isSubmitting ? 'Deleting...' : 'Delete Branch'}
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
