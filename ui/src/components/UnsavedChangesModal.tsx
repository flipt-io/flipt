import * as Dialog from '@radix-ui/react-dialog';
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
    <Dialog.Root open={isOpen} onOpenChange={onClose}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-20 bg-gray-500/75 dark:bg-gray-900/80 transition-opacity data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0" />
        <div className="fixed inset-0 z-20 flex items-center justify-center p-4">
          <Dialog.Content className="mx-auto max-w-lg rounded-lg bg-background dark:bg-gray-800 p-6 shadow-xl dark:shadow-2xl border dark:border-gray-700 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95">
            <Dialog.Title className="text-lg font-medium text-gray-900 dark:text-gray-100">
              Unsaved Changes
            </Dialog.Title>
            <Dialog.Description className="mt-2 text-sm text-muted-foreground">
              You have unsaved changes. Would you like to save them before
              leaving?
            </Dialog.Description>
            <div className="mt-8 flex justify-between">
              <Button variant="ghost" onClick={onClose}>
                Cancel
              </Button>
              <div className="space-x-3">
                <Button variant="secondary" onClick={onDiscard}>
                  Discard Changes
                </Button>
                <Button variant="primary" onClick={onSave}>
                  Save Changes
                </Button>
              </div>
            </div>
          </Dialog.Content>
        </div>
      </Dialog.Portal>
    </Dialog.Root>
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
    if (onDiscard) {
      onDiscard();
    } else {
      formik.resetForm();
    }
    setShowUnsavedModal(false);
    if (blocker.state === 'blocked') {
      blocker.proceed();
    } else {
      navigate(-1);
    }
  };

  const handleSave = async () => {
    if (onSave) {
      onSave();
    } else {
      await formik.submitForm();
    }
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
