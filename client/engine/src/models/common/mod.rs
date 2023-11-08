use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
pub enum FlagType {
    #[default]
    #[serde(rename = "VARIANT_FLAG_TYPE")]
    Variant,
    #[serde(rename = "BOOLEAN_FLAG_TYPE")]
    Boolean,
}

#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
pub enum SegmentOperator {
    #[default]
    #[serde(rename = "OR_SEGMENT_OPERATOR")]
    Or,
    #[serde(rename = "AND_SEGMENT_OPERATOR")]
    And,
}

#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
pub enum SegmentMatchType {
    #[default]
    #[serde(rename = "ANY_SEGMENT_MATCH_TYPE")]
    Any,
    #[serde(rename = "ALL_SEGMENT_MATCH_TYPE")]
    All,
}

#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
pub enum ConstraintComparisonType {
    #[default]
    #[serde(rename = "UNKNOWN_CONSTRAINT_COMPARISON_TYPE")]
    Unknown,
    #[serde(rename = "STRING_CONSTRAINT_COMPARISON_TYPE")]
    String,
    #[serde(rename = "NUMBER_CONSTRAINT_COMPARISON_TYPE")]
    Number,
    #[serde(rename = "BOOLEAN_CONSTRAINT_COMPARISON_TYPE")]
    Boolean,
    #[serde(rename = "DATETIME_CONSTRAINT_COMPARISON_TYPE")]
    DateTime,
}

#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
pub enum RolloutType {
    #[default]
    #[serde(rename = "UNKNOWN_ROLLOUT_TYPE")]
    Unknown,
    #[serde(rename = "SEGMENT_ROLLOUT_TYPE")]
    Segment,
    #[serde(rename = "THRESHOLD_ROLLOUT_TYPE")]
    Threshold,
}

#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub enum EvaluationReason {
    #[default]
    #[serde(rename = "UNKNOWN_EVALUATION_REASON")]
    Unknown,
    #[serde(rename = "FLAG_DISABLED_EVALUATION_REASON")]
    FlagDisabled,
    #[serde(rename = "MATCH_EVALUATION_REASON")]
    Match,
    #[serde(rename = "DEFAULT_EVALUATION_REASON")]
    Default,
}

#[derive(Clone, Debug, Deserialize, PartialEq)]
pub enum ErrorEvaluationReason {
    #[serde(rename = "UNKNOWN_ERROR_EVALUATION_REASON")]
    Unknown,
    #[serde(rename = "NOT_FOUND_ERROR_EVALUATION_REASON")]
    NotFound,
}
