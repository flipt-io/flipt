import { FormikProps } from 'formik';
import React, { ReactNode, createContext } from 'react';
import { v4 as uuid } from 'uuid';

import { IConstraint } from '~/types/Constraint';
import { ISegment } from '~/types/Segment';

interface SegmentFormContextProps {
  constraints: IConstraint[];
  updateSegment: (s: Partial<ISegment>) => void;
  setConstraints: (constraints: IConstraint[]) => void;
  createConstraint: (c: IConstraint) => void;
  updateConstraint: (c: IConstraint) => void;
  deleteConstraint: (c: IConstraint) => void;
}

// Define the context with default values
export const SegmentFormContext = createContext<SegmentFormContextProps>({
  constraints: [],
  updateSegment: () => {},
  setConstraints: () => {},
  createConstraint: () => {},
  updateConstraint: () => {},
  deleteConstraint: () => {}
});

interface SegmentFormProviderProps {
  children: ReactNode;
  formik: FormikProps<ISegment>;
}

export const SegmentFormProvider: React.FC<SegmentFormProviderProps> = ({
  children,
  formik
}) => {
  const { constraints } = formik.values;

  const updateSegment = (s: Partial<ISegment>) => {
    if ('name' in s) {
      formik.setFieldValue('name', s.name);
    }
    if ('description' in s) {
      formik.setFieldValue('description', s.description);
    }
    if ('matchType' in s) {
      formik.setFieldValue('matchType', s.matchType);
    }
  };

  const setConstraints = (constraints: IConstraint[]) => {
    formik.setFieldValue('constraints', constraints);
  };

  const createConstraint = (c: IConstraint) => {
    c.id = uuid();
    const newConstraints = [...(constraints || []), c];
    formik.setFieldValue('constraints', newConstraints);
  };

  const updateConstraint = (c: IConstraint) => {
    const newConstraints = [...(constraints || [])];
    const index = newConstraints.findIndex(
      (constraint) => constraint.id === c.id
    );
    newConstraints[index] = c;
    formik.setFieldValue('constraints', newConstraints);
  };

  const deleteConstraint = (c: IConstraint) => {
    const newConstraints = [...(constraints || [])];
    const index = newConstraints.findIndex(
      (constraint) => constraint.id === c.id
    );
    newConstraints.splice(index, 1);
    formik.setFieldValue('constraints', newConstraints);
  };

  return (
    <SegmentFormContext.Provider
      value={{
        constraints: constraints || [],
        updateSegment,
        setConstraints,
        createConstraint,
        updateConstraint,
        deleteConstraint
      }}
    >
      {children}
    </SegmentFormContext.Provider>
  );
};
