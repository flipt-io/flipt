import { Form, Formik } from 'formik';
import * as Yup from 'yup';

import {
  currentEnvironmentChanged,
  useDeleteBranchEnvironmentMutation
} from '~/app/environments/environmentsApi';

import Modal from '~/components/Modal';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';

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
        baseEnvironmentKey: baseBranch,
        environmentKey: environment.key
      });
      clearError();
      setSuccess('Branch deleted successfully');
      dispatch(currentEnvironmentChanged(baseBranch));
    } catch (e) {
      setError(e);
    }
  };
  return (
    <Modal open={open} setOpen={setOpen}>
      <div className="p-6 space-y-4">
        <h2 className="text-2xl font-bold">Delete Branch</h2>
        <p className="font-semibold">
          This action is <span className="underline">destructive</span> and
          cannot be undone.
        </p>
        <p>
          Deleting the branch{' '}
          <span className="font-mono font-semibold">{environment.key}</span>{' '}
          will permanently remove all of its data. It will also delete the
          branch from the remote repository if it exists.
        </p>
        <p>Please type the branch environment name to confirm.</p>
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
                <Input
                  name="confirmName"
                  type="text"
                  placeholder={environment.key}
                  className="w-full rounded-md border px-3 py-2 text-sm focus:ring-2 focus:ring-destructive focus:border-destructive transition mb-2"
                  disabled={isSubmitting}
                  onChange={formik.handleChange}
                  value={formik.values.confirmName}
                />
                {errors.confirmName && touched.confirmName && (
                  <div className="text-xs text-destructive mb-2">
                    {errors.confirmName}
                  </div>
                )}
                <div className="flex justify-end gap-2 mt-4">
                  <Button
                    variant="outline"
                    type="button"
                    onClick={() => setOpen(false)}
                    disabled={isSubmitting}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="destructive"
                    type="submit"
                    disabled={!isValid || isSubmitting}
                  >
                    {isSubmitting ? 'Deleting...' : 'Delete Branch'}
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
