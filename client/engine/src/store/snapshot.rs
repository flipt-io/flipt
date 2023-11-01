use super::models;
use crate::common;
use crate::flipt::models as flipt_models;
use snafu::{prelude::*, Whatever};
use std::collections::HashMap;

pub trait Parser {
    fn new(namespaces: Vec<String>) -> Self;
    fn parse(&self, namespace: String) -> Result<models::Document, Whatever>;
    fn get_namespaces(&self) -> Vec<String>;
}

pub struct Snapshot {
    namespace: HashMap<String, Namespace>,
}

struct Namespace {
    key: String,
    flags: HashMap<String, flipt_models::Flag>,
    eval_rules: HashMap<String, Vec<flipt_models::EvaluationRule>>,
    eval_rollouts: HashMap<String, Vec<flipt_models::EvaluationRollout>>,
    eval_distributions: HashMap<String, Vec<flipt_models::EvaluationDistribution>>,
}

impl Snapshot {
    pub fn build<T>(parser: T) -> Result<Snapshot, Whatever>
    where
        T: Parser,
    {
        let mut ns: HashMap<String, Namespace> = HashMap::new();

        for n in parser.get_namespaces() {
            let doc = parser.parse(n.clone())?;

            let mut flags: HashMap<String, flipt_models::Flag> = HashMap::new();
            let mut segments: HashMap<String, flipt_models::Segment> = HashMap::new();
            let mut eval_rules: HashMap<String, Vec<flipt_models::EvaluationRule>> = HashMap::new();
            let mut eval_rollouts: HashMap<String, Vec<flipt_models::EvaluationRollout>> =
                HashMap::new();
            let mut eval_dists: HashMap<String, Vec<flipt_models::EvaluationDistribution>> =
                HashMap::new();

            for segment in doc.segments {
                let mut s = flipt_models::Segment {
                    key: segment.key,
                    match_type: segment.match_type,
                    constraints: Vec::new(),
                };

                for constraint in segment.constraints {
                    let c = flipt_models::Constraint {
                        segment_key: s.key.clone(),
                        r#type: constraint.r#type,
                        property: constraint.property,
                        operator: constraint.operator,
                        value: constraint.value,
                    };

                    s.constraints.push(c);
                }

                segments.insert(s.key.clone(), s);
            }

            for flag in doc.flags {
                let mut f = flipt_models::Flag {
                    key: flag.key.clone(),
                    enabled: flag.enabled,
                    variants: Vec::new(),
                    r#type: flag.r#type.unwrap_or(common::FlagType::Variant),
                };

                let variants = flag.variants.unwrap_or(Vec::new());

                for variant in &variants {
                    // TODO(yquansah): Implement variant attachment support
                    f.variants.push(flipt_models::Variant {
                        key: variant.key.clone(),
                    })
                }

                flags.insert(f.key.clone(), f);

                // Flag Rules
                let mut eval_rules_collection: Vec<flipt_models::EvaluationRule> = Vec::new();
                let mut idx: usize = 0;

                let flag_rules = flag.rules.unwrap_or(vec![]);

                for rule in flag_rules {
                    idx += 1;

                    let rule_id = uuid::Uuid::new_v4().to_string();
                    let mut eval_rule = flipt_models::EvaluationRule {
                        id: rule_id.clone(),
                        rank: idx.clone(),
                        flag_key: flag.key.clone(),
                        segments: HashMap::new(),
                        segment_operator: common::SegmentOperator::Or,
                    };

                    let mut segment_keys: Vec<String> = Vec::new();
                    let mut segment_operator = common::SegmentOperator::Or;

                    if rule.segment.is_some() {
                        let segment_key = rule.segment.unwrap();
                        segment_keys.push(segment_key);
                    } else if rule.segments.is_some() {
                        let rule_segments = rule.segments.unwrap();

                        for segment_key in rule_segments.keys {
                            segment_keys.push(segment_key);
                        }

                        segment_operator = rule_segments.segment_operator;
                    }

                    for s in segment_keys {
                        let segment = match segments.get(&s) {
                            Some(s) => s,
                            None => whatever!("unknown segment reference: {}", s.clone()),
                        };

                        let mut eval_constraints: Vec<flipt_models::EvaluationConstraint> =
                            Vec::new();
                        for constraint in segment.constraints.iter() {
                            eval_constraints.push(flipt_models::EvaluationConstraint {
                                r#type: constraint.r#type.clone(),
                                property: constraint.property.clone(),
                                operator: constraint.operator.clone(),
                                value: constraint.value.clone(),
                            })
                        }

                        eval_rule.segment_operator = segment_operator.clone();

                        eval_rule.segments.insert(
                            s.clone(),
                            flipt_models::EvaluationSegment {
                                segment_key: s.clone(),
                                match_type: segment.match_type.clone(),
                                constraints: eval_constraints,
                            },
                        );
                    }

                    let mut evaluation_distributions: Vec<flipt_models::EvaluationDistribution> =
                        Vec::new();

                    for distribution in rule.distributions {
                        let variant_key = match find_by_key(&distribution.variant, &variants) {
                            Some(key) => key,
                            None => whatever!(
                                "unknown variant reference: {}",
                                distribution.variant.clone()
                            ),
                        };

                        evaluation_distributions.push(flipt_models::EvaluationDistribution {
                            rule_id: rule_id.clone(),
                            variant_key: variant_key,
                            rollout: distribution.rollout,
                        })
                    }

                    eval_dists.insert(rule_id.clone(), evaluation_distributions);

                    eval_rules_collection.push(eval_rule);
                }

                eval_rules.insert(flag.key.clone(), eval_rules_collection);

                // Flag Rollouts
                let mut eval_rollout_collection: Vec<flipt_models::EvaluationRollout> = Vec::new();
                let mut rollout_idx = 0;

                let flag_rollouts = flag.rollouts.unwrap_or(vec![]);

                for rollout in flag_rollouts {
                    rollout_idx += 1;

                    let mut evaluation_rollout: flipt_models::EvaluationRollout =
                        flipt_models::EvaluationRollout {
                            rank: rollout_idx,
                            rollout_type: common::RolloutType::Unknown,
                            segment: None,
                            threshold: None,
                        };

                    evaluation_rollout.rank = rollout_idx.clone();

                    if rollout.threshold.is_some() {
                        let threshold = rollout.threshold.unwrap();
                        evaluation_rollout.threshold = Some(flipt_models::RolloutThreshold {
                            percentage: threshold.percentage,
                            value: threshold.value,
                        });

                        evaluation_rollout.rollout_type = common::RolloutType::Threshold;
                    } else if rollout.segment.is_some() {
                        let mut evaluation_rollout_segments: HashMap<
                            String,
                            flipt_models::EvaluationSegment,
                        > = HashMap::new();

                        let segment_rule = rollout.segment.unwrap();

                        let mut segment_keys: Vec<String> = Vec::new();

                        if segment_rule.key.is_some() {
                            let s_key = segment_rule.key.unwrap();
                            if !s_key.is_empty() {
                                segment_keys.push(s_key);
                            }
                        } else if segment_rule.keys.is_some() {
                            let s_keys = segment_rule.keys.unwrap();

                            for key in s_keys {
                                segment_keys.push(key);
                            }
                        }

                        for segment_key in segment_keys {
                            let segment = match segments.get(&segment_key) {
                                Some(s) => s,
                                None => {
                                    whatever!("unknown segment reference: {}", segment_key.clone())
                                }
                            };

                            let mut constraints: Vec<flipt_models::EvaluationConstraint> =
                                Vec::new();
                            for constraint in &segment.constraints {
                                constraints.push(flipt_models::EvaluationConstraint {
                                    r#type: constraint.r#type.clone(),
                                    property: constraint.property.clone(),
                                    value: constraint.value.clone(),
                                    operator: constraint.operator.clone(),
                                })
                            }

                            evaluation_rollout_segments.insert(
                                segment.key.clone(),
                                flipt_models::EvaluationSegment {
                                    segment_key: segment.key.clone(),
                                    match_type: segment.match_type.clone(),
                                    constraints: constraints,
                                },
                            );
                        }

                        let segment_operator =
                            segment_rule.operator.unwrap_or(common::SegmentOperator::Or);

                        evaluation_rollout.rollout_type = common::RolloutType::Segment;
                        evaluation_rollout.segment = Some(flipt_models::RolloutSegment {
                            value: segment_rule.value,
                            segment_operator: segment_operator,
                            segments: evaluation_rollout_segments,
                        });
                    }

                    eval_rollout_collection.push(evaluation_rollout);
                }

                eval_rollouts.insert(flag.key.clone(), eval_rollout_collection);
            }

            ns.insert(
                n.clone(),
                Namespace {
                    key: n.clone(),
                    flags: flags,
                    eval_rules: eval_rules,
                    eval_rollouts: eval_rollouts,
                    eval_distributions: eval_dists,
                },
            );
        }

        Ok(Self { namespace: ns })
    }

    pub fn get_flag(&self, namespace_key: &str, flag_key: &str) -> Option<flipt_models::Flag> {
        let namespace = self.namespace.get(namespace_key)?;

        let flag = namespace.flags.get(flag_key)?;

        return Some(flag.clone());
    }

    pub fn get_evaluation_rules(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt_models::EvaluationRule>> {
        let namespace = self.namespace.get(namespace_key)?;

        let eval_rules = namespace.eval_rules.get(flag_key)?;

        return Some(eval_rules.to_vec());
    }

    pub fn get_evaluation_distributions(
        &self,
        namespace_key: &str,
        rule_id: &str,
    ) -> Option<Vec<flipt_models::EvaluationDistribution>> {
        let namespace = self.namespace.get(namespace_key)?;

        let evaluation_distributions = namespace.eval_distributions.get(rule_id)?;

        return Some(evaluation_distributions.to_vec());
    }

    pub fn get_evaluation_rollouts(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt_models::EvaluationRollout>> {
        let namespace = self.namespace.get(namespace_key)?;

        let eval_rollouts = namespace.eval_rollouts.get(flag_key)?;

        return Some(eval_rollouts.to_vec());
    }
}

fn find_by_key(key: &str, variants: &Vec<models::Variant>) -> Option<String> {
    match variants.iter().find(|&element| element.key == key) {
        Some(v) => Some(v.key.clone()),
        None => None,
    }
}

#[cfg(test)]
mod tests {
    use super::Parser;
    use super::Snapshot;
    use crate::common;
    use crate::flipt::models as flipt_models;
    use crate::flipt::models::EvaluationConstraint;
    use crate::store::parsers::TestParser;

    #[test]
    fn snapshot_tests() {
        let tp = TestParser::new(vec!["default".into()]);

        let snapshot = match Snapshot::build(tp) {
            Ok(s) => s,
            Err(e) => panic!("{}", e),
        };

        let flag_variant = snapshot
            .get_flag("default", "flag1")
            .expect("flag1 should exist");

        assert_eq!(flag_variant.key, "flag1");
        assert_eq!(flag_variant.variants.len(), 1);
        assert_eq!(flag_variant.enabled, true);
        assert_eq!(flag_variant.r#type, common::FlagType::Variant);

        let flag_boolean = snapshot
            .get_flag("default", "flag_boolean")
            .expect("flag_boolean should exist");

        assert_eq!(flag_boolean.key, "flag_boolean");
        assert_eq!(flag_boolean.enabled, true);
        assert_eq!(flag_boolean.r#type, common::FlagType::Boolean);

        let evaluation_rules = snapshot
            .get_evaluation_rules("default", "flag1")
            .expect("evaluation rules should exist for flag1");

        assert_eq!(evaluation_rules.len(), 1);
        assert_eq!(evaluation_rules[0].flag_key, "flag1");
        assert_eq!(
            evaluation_rules[0].segment_operator,
            common::SegmentOperator::Or
        );
        assert_eq!(evaluation_rules[0].rank, 1);
        assert_eq!(evaluation_rules[0].segments.len(), 1);
        assert_eq!(
            *evaluation_rules[0]
                .segments
                .get("segment1")
                .expect("segment1 should exist"),
            flipt_models::EvaluationSegment {
                segment_key: "segment1".into(),
                match_type: common::SegmentMatchType::Any,
                constraints: vec![EvaluationConstraint {
                    r#type: common::ConstraintComparisonType::String,
                    property: "fizz".into(),
                    operator: "eq".into(),
                    value: "buzz".into(),
                }],
            }
        );

        let evaluation_distributions = snapshot
            .get_evaluation_distributions("default", &evaluation_rules[0].id)
            .expect("evaluation distributions should exists for the rule");
        assert_eq!(evaluation_distributions.len(), 1);
        assert_eq!(evaluation_distributions[0].rollout, 100.0);
        assert_eq!(evaluation_distributions[0].rule_id, evaluation_rules[0].id);
        assert_eq!(evaluation_distributions[0].variant_key, "variant1");

        let evaluation_rollouts = snapshot
            .get_evaluation_rollouts("default", "flag_boolean")
            .expect("evaluation rollouts should exist for flag_boolean");

        assert_eq!(evaluation_rollouts.len(), 2);
        assert_eq!(evaluation_rollouts[0].rank, 1);
        assert_eq!(
            evaluation_rollouts[0].rollout_type,
            common::RolloutType::Segment
        );

        let segment_rollout = evaluation_rollouts[0]
            .segment
            .as_ref()
            .expect("first rollout should be segment");

        assert_eq!(segment_rollout.value, true);
        assert_eq!(
            segment_rollout.segment_operator,
            common::SegmentOperator::Or
        );
        assert_eq!(
            *segment_rollout
                .segments
                .get("segment1")
                .expect("segment1 should exist"),
            flipt_models::EvaluationSegment {
                segment_key: "segment1".into(),
                match_type: common::SegmentMatchType::Any,
                constraints: vec![EvaluationConstraint {
                    r#type: common::ConstraintComparisonType::String,
                    property: "fizz".into(),
                    operator: "eq".into(),
                    value: "buzz".into(),
                }],
            }
        );
    }
}
