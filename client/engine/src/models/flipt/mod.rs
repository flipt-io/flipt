use crate::models::common;
use serde::Deserialize;
use std::collections::HashMap;

#[derive(Clone)]
pub struct Flag {
    pub key: String,
    pub enabled: bool,
    pub r#type: common::FlagType,
}

#[derive(Clone, Deserialize)]
pub struct Variant {
    pub key: String,
    pub attachment: String,
}

#[derive(Clone)]
pub struct Constraint {
    pub segment_key: String,
    pub r#type: common::ConstraintComparisonType,
    pub property: String,
    pub operator: String,
    pub value: String,
}

#[derive(Clone, Debug)]
pub struct EvaluationRule {
    pub id: String,
    pub flag_key: String,
    pub segments: HashMap<String, EvaluationSegment>,
    pub rank: usize,
    pub segment_operator: common::SegmentOperator,
}

#[derive(Clone, Debug)]
pub struct EvaluationDistribution {
    pub rule_id: String,
    pub rollout: f32,
    pub variant_key: String,
    pub variant_attachment: String,
}

#[derive(Clone, Debug)]
pub struct EvaluationRollout {
    pub rollout_type: common::RolloutType,
    pub rank: usize,
    pub segment: Option<RolloutSegment>,
    pub threshold: Option<RolloutThreshold>,
}

#[derive(Clone, Debug)]
pub struct RolloutThreshold {
    pub percentage: f32,
    pub value: bool,
}

#[derive(Clone, Debug)]
pub struct RolloutSegment {
    pub value: bool,
    pub segment_operator: common::SegmentOperator,
    pub segments: HashMap<String, EvaluationSegment>,
}

#[derive(Clone, Debug, PartialEq)]
pub struct EvaluationSegment {
    pub segment_key: String,
    pub match_type: common::SegmentMatchType,
    pub constraints: Vec<EvaluationConstraint>,
}

#[derive(Clone, Debug, PartialEq)]
pub struct EvaluationConstraint {
    pub r#type: common::ConstraintComparisonType,
    pub property: String,
    pub operator: String,
    pub value: String,
}
