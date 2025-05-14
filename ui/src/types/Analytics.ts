export interface IFlagEvaluationCount {
  timestamps: string[];
  values: number[];
}

export interface IBatchFlagEvaluationCount {
  flagEvaluations: {
    [key: string]: IFlagEvaluationCount;
  };
}
