export interface IConstraintBase {
  property: string;
  type: ComparisonType;
  operator: string;
  value?: string;
  description?: string;
}

export interface IConstraint extends IConstraintBase {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export enum ComparisonType {
  STRING_COMPARISON_TYPE = 'string',
  NUMBER_COMPARISON_TYPE = 'number',
  BOOLEAN_COMPARISON_TYPE = 'boolean',
  DATETIME_COMPARISON_TYPE = 'datetime'
}

export const ConstraintStringOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  empty: 'IS EMPTY',
  notempty: 'IS NOT EMPTY',
  prefix: 'HAS PREFIX',
  suffix: 'HAS SUFFIX'
};

export const ConstraintNumberOperators: Record<string, string> = {
  eq: '==',
  neq: '!=',
  gt: '>',
  gte: '>=',
  lt: '<',
  lte: '<=',
  present: 'IS PRESENT',
  notpresent: 'IS NOT PRESENT'
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
  'empty',
  'notempty',
  'present',
  'notpresent'
];

export const ConstraintOperators: Record<string, string> = {
  ...ConstraintStringOperators,
  ...ConstraintNumberOperators,
  ...ConstraintBooleanOperators,
  ...ConstraintDateTimeOperators
};

export const toConstraintComparisonType = (t: string) =>
  ComparisonType[t as keyof typeof ComparisonType];

export const toConstraintOperator = (o: string) => ConstraintOperators[o];
