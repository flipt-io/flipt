import { FormikProps } from 'formik';
import React, { ReactNode, createContext } from 'react';
import { v4 as uuid } from 'uuid';

import { IFlag } from '~/types/Flag';
import { IRollout } from '~/types/Rollout';
import { IRule } from '~/types/Rule';
import { IVariant } from '~/types/Variant';

interface FlagFormContextProps {
  variants: IVariant[];
  rules: IRule[];
  rollouts: IRollout[];
  updateFlag: (f: Partial<IFlag>) => void;
  setRules: (rules: IRule[]) => void;
  createRule: (r: IRule) => void;
  updateRule: (r: IRule) => void;
  deleteRule: (r: IRule) => void;
  setVariants: (variants: IVariant[]) => void;
  createVariant: (v: IVariant) => void;
  updateVariant: (v: IVariant) => void;
  deleteVariant: (v: IVariant) => void;
  setRollouts: (rollouts: IRollout[]) => void;
  createRollout: (r: IRollout) => void;
  updateRollout: (r: IRollout) => void;
  deleteRollout: (r: IRollout) => void;
}

// Define the context with default values
export const FlagFormContext = createContext<FlagFormContextProps>({
  variants: [],
  rules: [],
  rollouts: [],
  updateFlag: () => {},
  setRules: () => {},
  createRule: () => {},
  updateRule: () => {},
  deleteRule: () => {},
  setVariants: () => {},
  createVariant: () => {},
  updateVariant: () => {},
  deleteVariant: () => {},
  setRollouts: () => {},
  createRollout: () => {},
  updateRollout: () => {},
  deleteRollout: () => {}
});

interface FlagFormProviderProps {
  children: ReactNode;
  formik: FormikProps<IFlag>;
}

export const FlagFormProvider: React.FC<FlagFormProviderProps> = ({
  children,
  formik
}) => {
  const { rules, variants, rollouts } = formik.values;

  const updateFlag = (f: Partial<IFlag>) => {
    if ('name' in f) {
      formik.setFieldValue('name', f.name);
    }
    if ('description' in f) {
      formik.setFieldValue('description', f.description);
    }
    if ('type' in f) {
      formik.setFieldValue('type', f.type);
    }
    if ('enabled' in f) {
      formik.setFieldValue('enabled', f.enabled);
    }
    if ('defaultVariant' in f) {
      formik.setFieldValue('defaultVariant', f.defaultVariant);
    }
  };

  const setRules = (rules: IRule[]) => {
    formik.setFieldValue('rules', rules);
  };

  const createRule = (r: IRule) => {
    r.id = uuid();
    const newRules = [...(rules || []), r];
    formik.setFieldValue('rules', newRules);
  };

  const updateRule = (r: IRule) => {
    const newRules = [...(rules || [])];
    const index = newRules.findIndex((rule) => rule.id === r.id);
    newRules[index] = r;
    formik.setFieldValue('rules', newRules);
  };

  const deleteRule = (r: IRule) => {
    const newRules = [...(rules || [])];
    const index = newRules.findIndex((rule) => rule.id === r.id);
    newRules.splice(index, 1);
    formik.setFieldValue('rules', newRules);
  };

  const setVariants = (variants: IVariant[]) => {
    formik.setFieldValue('variants', variants);
  };

  const createVariant = (v: IVariant) => {
    const newVariants = [...(variants || []), v];
    formik.setFieldValue('variants', newVariants);
  };

  const updateVariant = (v: IVariant) => {
    const newVariants = [...(variants || [])];
    const index = newVariants.findIndex((variant) => variant.key === v.key);
    newVariants[index] = v;
    formik.setFieldValue('variants', newVariants);
  };

  const deleteVariant = (v: IVariant) => {
    // first check if the variant is being used by a rule
    rules?.forEach((rule) => {
      if (rule.distributions.find((d) => d.variant === v.key)) {
        throw new Error('Cannot delete variant as it is being used by a rule');
      }
    });

    const newVariants = [...(variants || [])];
    const index = newVariants.findIndex((variant) => variant.key === v.key);
    newVariants.splice(index, 1);
    formik.setFieldValue('variants', newVariants);
  };

  const setRollouts = (rollouts: IRollout[]) => {
    formik.setFieldValue('rollouts', rollouts);
  };

  const createRollout = (r: IRollout) => {
    r.id = uuid();
    const newRollouts = [...(rollouts || []), r];
    formik.setFieldValue('rollouts', newRollouts);
  };

  const updateRollout = (r: IRollout) => {
    const newRollouts = [...(rollouts || [])];
    const index = newRollouts.findIndex((rollout) => rollout.id === r.id);
    newRollouts[index] = r;
    formik.setFieldValue('rollouts', newRollouts);
  };

  const deleteRollout = (r: IRollout) => {
    const newRollouts = [...(rollouts || [])];
    const index = newRollouts.findIndex((rollout) => rollout.id === r.id);
    newRollouts.splice(index, 1);
    formik.setFieldValue('rollouts', newRollouts);
  };

  return (
    <FlagFormContext.Provider
      value={{
        variants: variants || [],
        rules: rules || [],
        rollouts: rollouts || [],
        updateFlag,
        setRules,
        createRule,
        updateRule,
        deleteRule,
        setVariants,
        createVariant,
        updateVariant,
        deleteVariant,
        setRollouts,
        createRollout,
        updateRollout,
        deleteRollout
      }}
    >
      {children}
    </FlagFormContext.Provider>
  );
};
