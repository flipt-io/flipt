export interface IConstraintBase {
  property: string;
  type: ConstraintType;
  operator: string;
  value?: string;
  description?: string;
}

export interface IConstraint extends IConstraintBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export enum ConstraintType {
  STRING = 'STRING_COMPARISON_TYPE',
  NUMBER = 'NUMBER_COMPARISON_TYPE',
  BOOLEAN = 'BOOLEAN_COMPARISON_TYPE',
  DATETIME = 'DATETIME_COMPARISON_TYPE',
  ENTITY_ID = 'ENTITY_ID_COMPARISON_TYPE'
}

export function constraintTypeToLabel(c: ConstraintType): string {
  switch (c) {
    case ConstraintType.STRING:
      return 'String';
    case ConstraintType.NUMBER:
      return 'Number';
    case ConstraintType.BOOLEAN:
      return 'Boolean';
    case ConstraintType.DATETIME:
      return 'DateTime';
    case ConstraintType.ENTITY_ID:
      return 'Entity';
    default:
      return 'Unknown';
  }
}

export const ConstraintStringOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  empty: 'IS EMPTY',
  notempty: 'IS NOT EMPTY',
  prefix: 'HAS PREFIX',
  suffix: 'HAS SUFFIX',
  isoneof: 'IS ONE OF',
  isnotoneof: 'IS NOT ONE OF'
};

export const ConstraintEntityIdOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  isoneof: 'IS ONE OF',
  isnotoneof: 'IS NOT ONE OF'
};

export const ConstraintNumberOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  gt: '>',
  gte: '>=',
  lt: '<',
  lte: '<=',
  present: 'IS PRESENT',
  notpresent: 'IS NOT PRESENT',
  isoneof: 'IS ONE OF',
  isnotoneof: 'IS NOT ONE OF'
};

export const ConstraintBooleanOperators: Record<string, string> = {
  true: 'TRUE',
  false: 'FALSE',
  present: 'IS PRESENT',
  notpresent: 'IS NOT PRESENT'
};

export const ConstraintDateTimeOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  gt: 'IS AFTER',
  gte: 'IS AFTER OR AT',
  lt: 'IS BEFORE',
  lte: 'IS BEFORE OR AT',
  present: 'IS PRESENT',
  notpresent: 'IS NOT PRESENT'
};

export const NoValueOperators: string[] = [
  'true',
  'false',
  'empty',
  'notempty',
  'present',
  'notpresent'
];

export const ConstraintOperators: Record<string, string> = {
  ...ConstraintStringOperators,
  ...ConstraintNumberOperators,
  ...ConstraintBooleanOperators,
  ...ConstraintDateTimeOperators,
  ...ConstraintEntityIdOperators
};
