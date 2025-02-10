import { Dialog } from '@headlessui/react';
import { FormikProps } from 'formik';
import { useEffect, useState } from 'react';
import { useBlocker, useNavigate } from 'react-router';

import { Button } from './Button';

interface UnsavedChangesModalState {
  isOpen: boolean;
  onClose: () => void;
  onDiscard: () => void;
  onSave: () => void;
}

export function UnsavedChangesModal(props: UnsavedChangesModalState) {
  const { isOpen, onClose, onDiscard, onSave } = props;
  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <div className="fixed inset-0 bg-black/30" aria-hidden="true" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <Dialog.Panel className="mx-auto max-w-lg rounded-lg bg-white p-6 shadow-xl">
          <Dialog.Title className="text-lg font-medium text-gray-900">
            Unsaved Changes
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm text-gray-500">
            You have unsaved changes. Would you like to save them before
            leaving?
          </Dialog.Description>
          <div className="mt-4 flex justify-between">
            <Button onClick={onClose}>Cancel</Button>
            <div className="space-x-3">
              <Button variant="secondary" onClick={onDiscard}>
                Discard Changes
              </Button>
              <Button variant="primary" onClick={onSave}>
                Save Changes
              </Button>
            </div>
          </div>
        </Dialog.Panel>
      </div>
    </Dialog>
  );
}

interface UnsavedChangesModalProps {
  onDiscard?: () => void;
  onSave?: () => void;
  formik: FormikProps<any>;
  children: React.ReactNode;
}

export function UnsavedChangesModalWrapper(props: UnsavedChangesModalProps) {
  const { formik, children, onSave, onDiscard } = props;
  const [showUnsavedModal, setShowUnsavedModal] = useState(false);
  const navigate = useNavigate();

  // Block navigation when form is dirty
  const blocker = useBlocker(
    ({ currentLocation, nextLocation }) =>
      formik.dirty && currentLocation.pathname !== nextLocation.pathname
  );

  // Handle navigation attempts
  useEffect(() => {
    if (blocker.state === 'blocked') {
      setShowUnsavedModal(true);
    }
  }, [blocker]);

  // Add browser close/refresh guard
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (formik.dirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [formik.dirty]);

  const handleDiscard = () => {
    onDiscard ? onDiscard() : formik.resetForm();
    setShowUnsavedModal(false);
    if (blocker.state === 'blocked') {
      blocker.proceed();
    } else {
      navigate(-1);
    }
  };

  const handleSave = async () => {
    onSave ? onSave() : await formik.submitForm();
    setShowUnsavedModal(false);
    if (blocker.state === 'blocked') {
      blocker.proceed();
    }
  };

  return (
    <>
      {children}
      <UnsavedChangesModal
        isOpen={showUnsavedModal}
        onClose={() => {
          setShowUnsavedModal(false);
          if (blocker.state === 'blocked') {
            blocker.reset();
          }
        }}
        onDiscard={handleDiscard}
        onSave={handleSave}
      />
    </>
  );
}
