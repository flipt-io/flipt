use chrono::{DateTime, Utc};
use serde::Serialize;
use snafu::{prelude::*, Whatever};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use std::time::{Duration, SystemTime, SystemTimeError};

use crate::models::common;
use crate::models::flipt;
use crate::store::parsers;
use crate::store::snapshot;
use crate::store::snapshot::Parser;

const DEFAULT_PERCENT: f32 = 100.0;
const DEFAULT_TOTAL_BUCKET_NUMBER: i32 = 1000;
const DEFAULT_PERCENT_MULTIPIER: f32 = DEFAULT_TOTAL_BUCKET_NUMBER as f32 / DEFAULT_PERCENT;

pub struct Evaluator<T>
where
    T: snapshot::Parser,
{
    flipt_parser: T,
    snapshot: snapshot::Snapshot,
    mtx: Arc<RwLock<i32>>,
}

#[repr(C)]
pub struct EvaluationRequest {
    pub namespace_key: String,
    pub flag_key: String,
    pub entity_id: String,
    pub context: HashMap<String, String>,
}

#[derive(Serialize)]
pub struct VariantEvaluationResponse {
    pub r#match: bool,
    pub segment_keys: Vec<String>,
    pub reason: common::EvaluationReason,
    pub flag_key: String,
    pub variant_key: String,
    pub variant_attachment: String,
    pub request_duration_millis: f64,
    pub timestamp: DateTime<Utc>,
}

#[derive(Serialize)]
pub struct BooleanEvaluationResponse {
    pub enabled: bool,
    pub flag_key: String,
    pub reason: common::EvaluationReason,
    pub request_duration_millis: f64,
    pub timestamp: DateTime<Utc>,
}

pub struct ErrorEvaluationResponse {
    pub flag_key: String,
    pub reason: common::ErrorEvaluationReason,
}

pub trait EvaluationResponse {}

impl EvaluationResponse for VariantEvaluationResponse {}
impl EvaluationResponse for BooleanEvaluationResponse {}
impl EvaluationResponse for ErrorEvaluationResponse {}

impl Default for VariantEvaluationResponse {
    fn default() -> Self {
        Self {
            r#match: false,
            segment_keys: Vec::new(),
            reason: common::EvaluationReason::Unknown,
            flag_key: String::from(""),
            variant_key: String::from(""),
            variant_attachment: String::from(""),
            request_duration_millis: 0.0,
            timestamp: chrono::offset::Utc::now(),
        }
    }
}

impl Default for BooleanEvaluationResponse {
    fn default() -> Self {
        Self {
            enabled: false,
            flag_key: String::from(""),
            reason: common::EvaluationReason::Unknown,
            request_duration_millis: 0.0,
            timestamp: chrono::offset::Utc::now(),
        }
    }
}

impl Default for ErrorEvaluationResponse {
    fn default() -> Self {
        Self {
            flag_key: String::from(""),
            reason: common::ErrorEvaluationReason::Unknown,
        }
    }
}

type VariantEvaluationResult<T> = std::result::Result<T, Whatever>;

type BooleanEvaluationResult<T> = std::result::Result<T, Whatever>;

impl Evaluator<parsers::FliptParser> {
    pub fn new(namespaces: Vec<String>) -> Result<Self, Whatever> {
        let flipt_parser = parsers::FliptParser::new(namespaces.clone());
        let snap = snapshot::Snapshot::build(&flipt_parser)?;

        Ok(Self {
            flipt_parser,
            snapshot: snap,
            mtx: Arc::new(RwLock::new(0)),
        })
    }

    pub fn replace_snapshot(&mut self) {
        let _w_lock = self.mtx.write().unwrap();
        let snap = snapshot::Snapshot::build(&self.flipt_parser);
        self.snapshot = snap.unwrap();
    }

    pub fn variant(
        &self,
        evaluation_request: &EvaluationRequest,
    ) -> VariantEvaluationResult<VariantEvaluationResponse> {
        let _r_lock = self.mtx.read().unwrap();
        let flag = match self.snapshot.get_flag(
            &evaluation_request.namespace_key,
            &evaluation_request.flag_key,
        ) {
            Some(f) => {
                if f.r#type != common::FlagType::Variant {
                    whatever!("{} is not a variant flag", &evaluation_request.flag_key);
                }
                f
            }
            None => whatever!(
                "failed to get flag information {}/{}",
                &evaluation_request.namespace_key,
                &evaluation_request.flag_key
            ),
        };

        self.variant_evaluation(&flag, evaluation_request)
    }

    pub fn boolean(
        &self,
        evaluation_request: &EvaluationRequest,
    ) -> BooleanEvaluationResult<BooleanEvaluationResponse> {
        let _r_lock = self.mtx.read().unwrap();
        let flag = match self.snapshot.get_flag(
            &evaluation_request.namespace_key,
            &evaluation_request.flag_key,
        ) {
            Some(f) => {
                if f.r#type != common::FlagType::Boolean {
                    whatever!("{} is not a boolean flag", &evaluation_request.flag_key);
                }
                f
            }
            None => whatever!(
                "failed to get flag information {}/{}",
                &evaluation_request.namespace_key,
                &evaluation_request.flag_key
            ),
        };

        self.boolean_evaluation(&flag, evaluation_request)
    }

    pub fn batch(
        &self,
        requests: Vec<EvaluationRequest>,
    ) -> Result<Vec<Box<dyn EvaluationResponse>>, Whatever> {
        let mut evaluation_responses: Vec<Box<dyn EvaluationResponse>> = Vec::new();

        for request in requests {
            let flag = match self
                .snapshot
                .get_flag(&request.namespace_key, &request.flag_key)
            {
                Some(f) => f,
                None => {
                    evaluation_responses.push(Box::new(ErrorEvaluationResponse {
                        flag_key: request.flag_key.clone(),
                        reason: common::ErrorEvaluationReason::NotFound,
                    }));
                    continue;
                }
            };

            match flag.r#type {
                common::FlagType::Boolean => {
                    match self.boolean_evaluation(&flag, &request) {
                        Ok(b) => {
                            evaluation_responses.push(Box::new(b));
                        }
                        Err(e) => {
                            return Err(e);
                        }
                    };
                }
                common::FlagType::Variant => {
                    match self.variant_evaluation(&flag, &request) {
                        Ok(v) => {
                            evaluation_responses.push(Box::new(v));
                        }
                        Err(e) => {
                            return Err(e);
                        }
                    };
                }
            }
        }

        Ok(evaluation_responses)
    }

    fn variant_evaluation(
        &self,
        flag: &flipt::Flag,
        evaluation_request: &EvaluationRequest,
    ) -> VariantEvaluationResult<VariantEvaluationResponse> {
        let now = SystemTime::now();
        let mut last_rank = 0;

        let mut variant_evaluation_response = VariantEvaluationResponse {
            flag_key: flag.key.clone(),
            ..Default::default()
        };

        if !flag.enabled {
            variant_evaluation_response.reason = common::EvaluationReason::FlagDisabled;
            variant_evaluation_response.request_duration_millis =
                get_duration_millis(now.elapsed())?;
            return Ok(variant_evaluation_response);
        }

        let evaluation_rules = match self.snapshot.get_evaluation_rules(
            &evaluation_request.namespace_key,
            &evaluation_request.flag_key,
        ) {
            Some(evaluation_rules) => evaluation_rules,
            None => whatever!(
                "error getting evaluation rules for namespace {} and flag {}",
                evaluation_request.namespace_key.clone(),
                evaluation_request.flag_key.clone()
            ),
        };

        for rule in evaluation_rules {
            if rule.rank < last_rank {
                whatever!("rule rank: {} detected out of order", rule.rank);
            }

            last_rank = rule.rank;

            let mut segment_keys: Vec<String> = Vec::new();
            let mut segment_matches = 0;

            for kv in &rule.segments {
                let matched = match self.matches_constraints(
                    &evaluation_request.context,
                    &kv.1.constraints,
                    &kv.1.match_type,
                ) {
                    Ok(b) => b,
                    Err(err) => return Err(err),
                };

                if matched {
                    segment_keys.push(kv.0.into());
                    segment_matches += 1;
                }
            }

            if rule.segment_operator == common::SegmentOperator::Or {
                if segment_matches < 1 {
                    continue;
                }
            } else if rule.segment_operator == common::SegmentOperator::And
                && rule.segments.len() != segment_matches
            {
                continue;
            }

            variant_evaluation_response.segment_keys = segment_keys;

            let distributions = match self
                .snapshot
                .get_evaluation_distributions(&evaluation_request.namespace_key, &rule.id)
            {
                Some(evaluation_distributions) => evaluation_distributions,
                None => whatever!(
                    "error getting evaluation distributions for namespace {} and rule {}",
                    evaluation_request.namespace_key.clone(),
                    rule.id.clone()
                ),
            };

            let mut valid_distributions: Vec<flipt::EvaluationDistribution> = Vec::new();
            let mut buckets: Vec<i32> = Vec::new();

            for distribution in distributions {
                if distribution.rollout > 0.0 {
                    valid_distributions.push(distribution.clone());
                }

                if buckets.is_empty() {
                    let bucket = (distribution.rollout * DEFAULT_PERCENT_MULTIPIER) as i32;
                    buckets.push(bucket);
                } else {
                    let bucket = buckets[buckets.len() - 1]
                        + (distribution.rollout * DEFAULT_PERCENT_MULTIPIER) as i32;
                    buckets.push(bucket);
                }
            }

            // no distributions for the rule
            if valid_distributions.is_empty() {
                variant_evaluation_response.r#match = true;
                variant_evaluation_response.reason = common::EvaluationReason::Match;
                variant_evaluation_response.request_duration_millis =
                    get_duration_millis(now.elapsed())?;
                return Ok(variant_evaluation_response);
            }

            let bucket = crc32fast::hash(
                format!(
                    "{}{}",
                    evaluation_request.entity_id, evaluation_request.flag_key
                )
                .as_bytes(),
            ) as i32
                % DEFAULT_TOTAL_BUCKET_NUMBER;

            buckets.sort();

            let index = match buckets.binary_search(&bucket) {
                Ok(idx) => idx,
                Err(idx) => idx,
            };

            if index == valid_distributions.len() {
                variant_evaluation_response.r#match = false;
                variant_evaluation_response.request_duration_millis =
                    get_duration_millis(now.elapsed())?;
                return Ok(variant_evaluation_response);
            }

            let d = &valid_distributions[index];

            variant_evaluation_response.r#match = true;
            variant_evaluation_response.variant_key = d.variant_key.clone();
            variant_evaluation_response.variant_attachment = d.variant_attachment.clone();
            variant_evaluation_response.reason = common::EvaluationReason::Match;
            variant_evaluation_response.request_duration_millis =
                get_duration_millis(now.elapsed())?;
            return Ok(variant_evaluation_response);
        }

        Ok(variant_evaluation_response)
    }

    fn boolean_evaluation(
        &self,
        flag: &flipt::Flag,
        evaluation_request: &EvaluationRequest,
    ) -> BooleanEvaluationResult<BooleanEvaluationResponse> {
        let now = SystemTime::now();
        let mut last_rank = 0;

        let evaluation_rollouts = match self.snapshot.get_evaluation_rollouts(
            &evaluation_request.namespace_key,
            &evaluation_request.flag_key,
        ) {
            Some(rollouts) => rollouts,
            None => whatever!(
                "error getting evaluation rollouts for namespace {} and flag {}",
                evaluation_request.namespace_key.clone(),
                evaluation_request.flag_key.clone()
            ),
        };

        for rollout in evaluation_rollouts {
            if rollout.rank < last_rank {
                whatever!("rollout rank: {} detected out of order", rollout.rank);
            }

            last_rank = rollout.rank;

            if rollout.threshold.is_some() {
                let threshold = rollout.threshold.unwrap();

                let normalized_value = (crc32fast::hash(
                    format!(
                        "{}{}",
                        evaluation_request.entity_id, evaluation_request.flag_key
                    )
                    .as_bytes(),
                ) as i32
                    % 100) as f32;

                if normalized_value < threshold.percentage {
                    return Ok(BooleanEvaluationResponse {
                        enabled: threshold.value,
                        flag_key: flag.key.clone(),
                        reason: common::EvaluationReason::Match,
                        request_duration_millis: get_duration_millis(now.elapsed())?,
                        timestamp: chrono::offset::Utc::now(),
                    });
                }
            } else if rollout.segment.is_some() {
                let segment = rollout.segment.unwrap();
                let mut segment_matches = 0;

                for s in &segment.segments {
                    let matched = match self.matches_constraints(
                        &evaluation_request.context,
                        &s.1.constraints,
                        &s.1.match_type,
                    ) {
                        Ok(v) => v,
                        Err(err) => return Err(err),
                    };

                    if matched {
                        segment_matches += 1;
                    }
                }

                if segment.segment_operator == common::SegmentOperator::Or {
                    if segment_matches < 1 {
                        continue;
                    }
                } else if segment.segment_operator == common::SegmentOperator::And
                    && segment.segments.len() != segment_matches
                {
                    continue;
                }

                return Ok(BooleanEvaluationResponse {
                    enabled: segment.value,
                    flag_key: flag.key.clone(),
                    reason: common::EvaluationReason::Match,
                    request_duration_millis: get_duration_millis(now.elapsed())?,
                    timestamp: chrono::offset::Utc::now(),
                });
            }
        }

        Ok(BooleanEvaluationResponse {
            enabled: flag.enabled,
            flag_key: flag.key.clone(),
            reason: common::EvaluationReason::Default,
            request_duration_millis: get_duration_millis(now.elapsed())?,
            timestamp: chrono::offset::Utc::now(),
        })
    }

    fn matches_constraints(
        &self,
        eval_context: &HashMap<String, String>,
        constraints: &Vec<flipt::EvaluationConstraint>,
        segment_match_type: &common::SegmentMatchType,
    ) -> Result<bool, Whatever> {
        let mut constraint_matches: usize = 0;

        for constraint in constraints {
            let value = match eval_context.get(&constraint.property) {
                Some(v) => v,
                None => continue,
            };

            let matched = match constraint.r#type {
                common::ConstraintComparisonType::String => matches_string(constraint, value),
                common::ConstraintComparisonType::Number => matches_number(constraint, value)?,
                common::ConstraintComparisonType::Boolean => matches_boolean(constraint, value)?,
                common::ConstraintComparisonType::DateTime => matches_datetime(constraint, value)?,
                _ => {
                    return Ok(false);
                }
            };

            if matched {
                constraint_matches += 1;

                if segment_match_type == &common::SegmentMatchType::Any {
                    break;
                } else {
                    continue;
                }
            } else if segment_match_type == &common::SegmentMatchType::All {
                break;
            } else {
                continue;
            }
        }

        let is_match = match segment_match_type {
            common::SegmentMatchType::All => constraints.len() == constraint_matches,
            common::SegmentMatchType::Any => constraints.is_empty() || constraint_matches != 0,
        };

        Ok(is_match)
    }
}

fn matches_string(evaluation_constraint: &flipt::EvaluationConstraint, v: &str) -> bool {
    let operator = evaluation_constraint.operator.as_str();

    match operator {
        "empty" => {
            return v.is_empty();
        }
        "notempty" => {
            return !v.is_empty();
        }
        _ => {}
    }

    if v.is_empty() {
        return false;
    }

    let value = evaluation_constraint.value.as_str();
    match operator {
        "eq" => v == value,
        "neq" => v != value,
        "prefix" => v.starts_with(value),
        "suffix" => v.ends_with(value),
        _ => false,
    }
}

fn matches_number(
    evaluation_constraint: &flipt::EvaluationConstraint,
    v: &str,
) -> Result<bool, Whatever> {
    let operator = evaluation_constraint.operator.as_str();

    match operator {
        "notpresent" => {
            return Ok(v.is_empty());
        }
        "present" => {
            return Ok(!v.is_empty());
        }
        _ => {}
    }

    if v.is_empty() {
        return Ok(false);
    }

    let v_number = match v.parse::<i32>() {
        Ok(v) => v,
        Err(err) => whatever!("error parsing number {}, err: {}", v, err),
    };

    let value_number = match evaluation_constraint.value.parse::<i32>() {
        Ok(v) => v,
        Err(err) => whatever!(
            "error parsing number {}, err: {}",
            evaluation_constraint.value,
            err
        ),
    };

    match operator {
        "eq" => Ok(v_number == value_number),
        "neq" => Ok(v_number != value_number),
        "lt" => Ok(v_number < value_number),
        "lte" => Ok(v_number <= value_number),
        "gt" => Ok(v_number > value_number),
        "gte" => Ok(v_number >= value_number),
        _ => Ok(false),
    }
}

fn matches_boolean(
    evaluation_constraint: &flipt::EvaluationConstraint,
    v: &str,
) -> Result<bool, Whatever> {
    let operator = evaluation_constraint.operator.as_str();

    match operator {
        "notpresent" => {
            return Ok(v.is_empty());
        }
        "present" => {
            return Ok(!v.is_empty());
        }
        _ => {}
    }

    if v.is_empty() {
        return Ok(false);
    }

    let v_bool = match v.parse::<bool>() {
        Ok(v) => v,
        Err(err) => whatever!("error parsing boolean {}: err {}", v, err),
    };

    match operator {
        "true" => Ok(v_bool),
        "false" => Ok(!v_bool),
        _ => Ok(false),
    }
}

fn matches_datetime(
    evaluation_constraint: &flipt::EvaluationConstraint,
    v: &str,
) -> Result<bool, Whatever> {
    let operator = evaluation_constraint.operator.as_str();

    match operator {
        "notpresent" => {
            return Ok(v.is_empty());
        }
        "present" => {
            return Ok(!v.is_empty());
        }
        _ => {}
    }

    if v.is_empty() {
        return Ok(false);
    }

    let d = match DateTime::parse_from_rfc3339(v) {
        Ok(t) => t.timestamp(),
        Err(e) => whatever!("error parsing time {}, err: {}", v, e),
    };

    let value = match DateTime::parse_from_rfc3339(&evaluation_constraint.value) {
        Ok(t) => t.timestamp(),
        Err(e) => whatever!(
            "error parsing time {}, err: {}",
            &evaluation_constraint.value,
            e
        ),
    };

    match operator {
        "eq" => Ok(d == value),
        "neq" => Ok(d != value),
        "lt" => Ok(d < value),
        "lte" => Ok(d <= value),
        "gt" => Ok(d > value),
        "gte" => Ok(d >= value),
        _ => Ok(false),
    }
}

fn get_duration_millis(elapsed: Result<Duration, SystemTimeError>) -> Result<f64, Whatever> {
    match elapsed {
        Ok(elapsed) => Ok(elapsed.as_secs_f64() * 1000.0),
        Err(e) => {
            whatever!("error getting duration {}", e)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::{matches_boolean, matches_datetime, matches_number, matches_string};
    use crate::models::common;
    use crate::models::flipt;

    macro_rules! matches_string_tests {
        ($($name:ident: $value:expr,)*) => {
        $(
            #[test]
            fn $name() {
                let (first, second, expected) = $value;
                assert_eq!(expected, matches_string(first, second));
            }
        )*
        }
    }

    macro_rules! matches_datetime_tests {
        ($($name:ident: $value:expr,)*) => {
        $(
            #[test]
            fn $name() {
                let (first, second, expected) = $value;
                assert_eq!(expected, matches_datetime(first, second).unwrap());
            }
        )*
        }
    }

    macro_rules! matches_number_tests {
        ($($name:ident: $value:expr,)*) => {
        $(
            #[test]
            fn $name() {
                let (first, second, expected) = $value;
                assert_eq!(expected, matches_number(first, second).unwrap());
            }
        )*
        }
    }

    matches_string_tests! {
        string_eq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::String,
            property: String::from("number"),
            operator: String::from("eq"),
            value: String::from("number"),
        }, "number", true),
        string_neq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::String,
            property: String::from("number"),
            operator: String::from("neq"),
            value: String::from("number"),
        }, "num", true),
        string_prefix: (&flipt::EvaluationConstraint{
                r#type: common::ConstraintComparisonType::String,
                property: String::from("number"),
                operator: String::from("prefix"),
                value: String::from("num"),
            }, "number", true),
        string_suffix: (&flipt::EvaluationConstraint{
                r#type: common::ConstraintComparisonType::String,
                property: String::from("number"),
                operator: String::from("suffix"),
                value: String::from("ber"),
            }, "number", true),
    }

    matches_datetime_tests! {
        datetime_eq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("eq"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T15:04:05Z", true),
        datetime_neq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("neq"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T15:03:05Z", true),
        datetime_lt: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("lt"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T14:03:05Z", true),
        datetime_gt: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("gt"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T16:03:05Z", true),
        datetime_lte: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("lte"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T15:04:05Z", true),
        datetime_gte: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::DateTime,
            property: String::from("date"),
            operator: String::from("gte"),
            value: String::from("2006-01-02T15:04:05Z"),
        }, "2006-01-02T16:03:05Z", true),

    }

    matches_number_tests! {
        number_eq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("number"),
            operator: String::from("eq"),
            value: String::from("1"),
        }, "1", true),
        number_neq: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("number"),
            operator: String::from("neq"),
            value: String::from("1"),
        }, "0", true),
        number_lt: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("number"),
            operator: String::from("lt"),
            value: String::from("4"),
        }, "3", true),
        number_gt: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("number"),
            operator: String::from("gt"),
            value: String::from("3"),
        }, "4", true),
        number_lte: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("date"),
            operator: String::from("lte"),
            value: String::from("3"),
        }, "3", true),
        number_gte: (&flipt::EvaluationConstraint{
            r#type: common::ConstraintComparisonType::Number,
            property: String::from("date"),
            operator: String::from("gte"),
            value: String::from("3"),
        }, "4", true),

    }

    #[test]
    fn test_matches_boolean_success() {
        let value_one = matches_boolean(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::Boolean,
                property: String::from("fizz"),
                operator: String::from("true"),
                value: "".into(),
            },
            "true",
        )
        .expect("boolean should be parsed correctly");

        assert!(value_one);

        let value_two = matches_boolean(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::Boolean,
                property: String::from("fizz"),
                operator: String::from("false"),
                value: "".into(),
            },
            "false",
        )
        .expect("boolean should be parsed correctly");

        assert!(value_two);
    }

    #[test]
    fn test_matches_boolean_failure() {
        let result = matches_boolean(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::Boolean,
                property: String::from("fizz"),
                operator: String::from("true"),
                value: "".into(),
            },
            "blah",
        );

        assert!(!result.is_ok());
        assert_eq!(
            result.err().unwrap().to_string(),
            "error parsing boolean blah: err provided string was not `true` or `false`"
        );
    }

    #[test]
    fn test_matches_number_failure() {
        let result_one = matches_number(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::Number,
                property: String::from("number"),
                operator: String::from("eq"),
                value: String::from("9"),
            },
            "notanumber",
        );

        assert!(!result_one.is_ok());
        assert_eq!(
            result_one.err().unwrap().to_string(),
            "error parsing number notanumber, err: invalid digit found in string"
        );

        let result_two = matches_number(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::Number,
                property: String::from("number"),
                operator: String::from("eq"),
                value: String::from("notanumber"),
            },
            "9",
        );

        assert!(!result_two.is_ok());
        assert_eq!(
            result_two.err().unwrap().to_string(),
            "error parsing number notanumber, err: invalid digit found in string"
        );
    }

    #[test]
    fn test_matches_datetime_failure() {
        let result_one = matches_datetime(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::String,
                property: String::from("date"),
                operator: String::from("eq"),
                value: String::from("blah"),
            },
            "2006-01-02T15:04:05Z",
        );

        assert!(!result_one.is_ok());
        assert_eq!(
            result_one.err().unwrap().to_string(),
            "error parsing time blah, err: input contains invalid characters"
        );

        let result_two = matches_datetime(
            &flipt::EvaluationConstraint {
                r#type: common::ConstraintComparisonType::String,
                property: String::from("date"),
                operator: String::from("eq"),
                value: String::from("2006-01-02T15:04:05Z"),
            },
            "blah",
        );

        assert!(!result_two.is_ok());
        assert_eq!(
            result_two.err().unwrap().to_string(),
            "error parsing time blah, err: input contains invalid characters"
        );
    }
}
