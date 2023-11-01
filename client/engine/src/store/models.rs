use crate::common;
use serde::Deserialize;

#[derive(Deserialize)]
pub struct Document {
    pub version: String,
    pub namespace: String,
    pub flags: Vec<Flag>,
    pub segments: Vec<Segment>,
}

#[derive(Deserialize)]
pub struct Flag {
    pub key: String,
    pub name: String,
    pub r#type: Option<common::FlagType>,
    pub description: Option<String>,
    pub enabled: bool,
    pub variants: Option<Vec<Variant>>,
    pub rules: Option<Vec<Rule>>,
    pub rollouts: Option<Vec<Rollout>>,
}

#[derive(Deserialize)]
pub struct Variant {
    pub key: String,
    pub name: String,
    pub description: String,
}

#[derive(Deserialize)]
pub struct Rule {
    pub distributions: Vec<Distribution>,
    pub segment: Option<String>,
    pub segments: Option<Segments>,
}

#[derive(Deserialize)]
pub struct Distribution {
    pub variant: String,
    pub rollout: f32,
}

#[derive(Deserialize)]
pub struct Rollout {
    pub description: Option<String>,
    pub segment: Option<SegmentRule>,
    pub threshold: Option<ThresholdRule>,
}

#[derive(Deserialize)]
pub struct SegmentRule {
    pub key: Option<String>,
    pub keys: Option<Vec<String>>,
    pub operator: Option<common::SegmentOperator>,
    pub value: bool,
}

#[derive(Deserialize)]
pub struct ThresholdRule {
    pub percentage: f32,
    pub value: bool,
}

#[derive(Deserialize)]
pub struct Segment {
    pub key: String,
    pub name: String,
    pub description: String,
    pub match_type: common::SegmentMatchType,
    pub constraints: Vec<Constraint>,
}

#[derive(Deserialize)]
pub struct Constraint {
    pub r#type: common::ConstraintComparisonType,
    pub property: String,
    pub operator: String,
    pub value: String,
    pub description: Option<String>,
}

#[derive(Deserialize)]
pub struct Segments {
    pub keys: Vec<String>,
    pub segment_operator: common::SegmentOperator,
}
