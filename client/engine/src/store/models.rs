use crate::common;
use serde::Deserialize;

#[derive(Deserialize)]
pub struct Document {
    pub namespace: Namespace,
    pub flags: Vec<Flag>,
}

#[derive(Deserialize)]
pub struct Namespace {
    pub key: String,
    pub name: String,
}

#[derive(Deserialize)]
pub struct Flag {
    pub key: String,
    pub name: String,
    pub r#type: Option<common::FlagType>,
    pub description: Option<String>,
    pub enabled: bool,
    pub rules: Option<Vec<Rule>>,
    pub rollouts: Option<Vec<Rollout>>,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Rule {
    pub distributions: Vec<Distribution>,
    pub segments: Option<Vec<Segment>>,
    pub segment_operator: common::SegmentOperator,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Distribution {
    pub variant_key: String,
    pub rollout: f32,
    pub variant_attachment: String,
}

#[derive(Deserialize)]
pub struct Rollout {
    pub description: Option<String>,
    pub segment: Option<SegmentRule>,
    pub threshold: Option<Threshold>,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SegmentRule {
    pub segment_operator: Option<common::SegmentOperator>,
    pub value: bool,
    pub segments: Vec<Segment>,
}

#[derive(Deserialize)]
pub struct Threshold {
    pub percentage: f32,
    pub value: bool,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Segment {
    pub key: String,
    pub match_type: common::SegmentMatchType,
    pub constraints: Vec<SegmentConstraint>,
}

#[derive(Deserialize)]
pub struct SegmentConstraint {
    pub r#type: common::ConstraintComparisonType,
    pub property: String,
    pub operator: String,
    pub value: String,
}
