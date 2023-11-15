use crate::models::common;
use crate::models::flipt;
use crate::models::transport;

#[cfg(test)]
use mockall::automock;
use snafu::Whatever;
use std::collections::HashMap;

pub trait Parser {
    fn parse(&self, namespace: String) -> Result<transport::Document, Whatever>;
    fn get_namespaces(&self) -> Vec<String>;
}

#[cfg_attr(test, automock)]
pub trait Store {
    fn get_flag(&self, namespace_key: &str, flag_key: &str) -> Option<flipt::Flag>;
    fn get_evaluation_rules(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt::EvaluationRule>>;
    fn get_evaluation_distributions(
        &self,
        namespace_key: &str,
        rule_id: &str,
    ) -> Option<Vec<flipt::EvaluationDistribution>>;
    fn get_evaluation_rollouts(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt::EvaluationRollout>>;
}

pub struct Snapshot {
    namespace: HashMap<String, Namespace>,
}

struct Namespace {
    _key: String,
    flags: HashMap<String, flipt::Flag>,
    eval_rules: HashMap<String, Vec<flipt::EvaluationRule>>,
    eval_rollouts: HashMap<String, Vec<flipt::EvaluationRollout>>,
    eval_distributions: HashMap<String, Vec<flipt::EvaluationDistribution>>,
}

impl Snapshot {
    pub fn build(parser: &dyn Parser) -> Result<Snapshot, Whatever> {
        let mut ns: HashMap<String, Namespace> = HashMap::new();

        for n in parser.get_namespaces() {
            let doc = parser.parse(n.clone())?;

            let mut flags: HashMap<String, flipt::Flag> = HashMap::new();
            let mut eval_rules: HashMap<String, Vec<flipt::EvaluationRule>> = HashMap::new();
            let mut eval_rollouts: HashMap<String, Vec<flipt::EvaluationRollout>> = HashMap::new();
            let mut eval_dists: HashMap<String, Vec<flipt::EvaluationDistribution>> =
                HashMap::new();

            for flag in doc.flags {
                let f = flipt::Flag {
                    key: flag.key.clone(),
                    enabled: flag.enabled,
                    r#type: flag.r#type.unwrap_or(common::FlagType::Variant),
                };

                flags.insert(f.key.clone(), f);

                // Flag Rules
                let mut eval_rules_collection: Vec<flipt::EvaluationRule> = Vec::new();

                let flag_rules = flag.rules.unwrap_or(vec![]);

                for (idx, rule) in flag_rules.into_iter().enumerate() {
                    let rule_id = uuid::Uuid::new_v4().to_string();
                    let mut eval_rule = flipt::EvaluationRule {
                        id: rule_id.clone(),
                        rank: idx + 1,
                        flag_key: flag.key.clone(),
                        segments: HashMap::new(),
                        segment_operator: rule.segment_operator,
                    };

                    if rule.segments.is_some() {
                        let rule_segments = rule.segments.unwrap();

                        for rule_segment in rule_segments {
                            let mut eval_constraints: Vec<flipt::EvaluationConstraint> = Vec::new();
                            for constraint in rule_segment.constraints {
                                eval_constraints.push(flipt::EvaluationConstraint {
                                    r#type: constraint.r#type,
                                    property: constraint.property,
                                    operator: constraint.operator,
                                    value: constraint.value,
                                });
                            }

                            eval_rule.segments.insert(
                                rule_segment.key.clone(),
                                flipt::EvaluationSegment {
                                    segment_key: rule_segment.key,
                                    match_type: rule_segment.match_type,
                                    constraints: eval_constraints,
                                },
                            );
                        }
                    }

                    let mut evaluation_distributions: Vec<flipt::EvaluationDistribution> =
                        Vec::new();

                    for distribution in rule.distributions {
                        evaluation_distributions.push(flipt::EvaluationDistribution {
                            rule_id: rule_id.clone(),
                            variant_key: distribution.variant_key,
                            variant_attachment: distribution.variant_attachment,
                            rollout: distribution.rollout,
                        })
                    }

                    eval_dists.insert(rule_id.clone(), evaluation_distributions);

                    eval_rules_collection.push(eval_rule);
                }

                eval_rules.insert(flag.key.clone(), eval_rules_collection);

                // Flag Rollouts
                let mut eval_rollout_collection: Vec<flipt::EvaluationRollout> = Vec::new();
                let mut rollout_idx = 0;

                let flag_rollouts = flag.rollouts.unwrap_or(vec![]);

                for rollout in flag_rollouts {
                    rollout_idx += 1;

                    let mut evaluation_rollout: flipt::EvaluationRollout =
                        flipt::EvaluationRollout {
                            rank: rollout_idx,
                            rollout_type: common::RolloutType::Unknown,
                            segment: None,
                            threshold: None,
                        };

                    evaluation_rollout.rank = rollout_idx;

                    if rollout.threshold.is_some() {
                        let threshold = rollout.threshold.unwrap();
                        evaluation_rollout.threshold = Some(flipt::RolloutThreshold {
                            percentage: threshold.percentage,
                            value: threshold.value,
                        });

                        evaluation_rollout.rollout_type = common::RolloutType::Threshold;
                    } else if rollout.segment.is_some() {
                        let mut evaluation_rollout_segments: HashMap<
                            String,
                            flipt::EvaluationSegment,
                        > = HashMap::new();

                        let segment_rule = rollout.segment.unwrap();

                        for segment in segment_rule.segments {
                            let mut constraints: Vec<flipt::EvaluationConstraint> = Vec::new();
                            for constraint in segment.constraints {
                                constraints.push(flipt::EvaluationConstraint {
                                    r#type: constraint.r#type,
                                    property: constraint.property,
                                    value: constraint.value,
                                    operator: constraint.operator,
                                });
                            }

                            evaluation_rollout_segments.insert(
                                segment.key.clone(),
                                flipt::EvaluationSegment {
                                    segment_key: segment.key,
                                    match_type: segment.match_type.clone(),
                                    constraints,
                                },
                            );
                        }

                        evaluation_rollout.rollout_type = common::RolloutType::Segment;
                        evaluation_rollout.segment = Some(flipt::RolloutSegment {
                            value: segment_rule.value,
                            segment_operator: segment_rule
                                .segment_operator
                                .unwrap_or(common::SegmentOperator::Or),
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
                    _key: n.clone(),
                    flags,
                    eval_rules,
                    eval_rollouts,
                    eval_distributions: eval_dists,
                },
            );
        }

        Ok(Self { namespace: ns })
    }
}

impl Store for Snapshot {
    fn get_flag(&self, namespace_key: &str, flag_key: &str) -> Option<flipt::Flag> {
        let namespace = self.namespace.get(namespace_key)?;

        let flag = namespace.flags.get(flag_key)?;

        Some(flag.clone())
    }

    fn get_evaluation_rules(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt::EvaluationRule>> {
        let namespace = self.namespace.get(namespace_key)?;

        let eval_rules = namespace.eval_rules.get(flag_key)?;

        Some(eval_rules.to_vec())
    }

    fn get_evaluation_distributions(
        &self,
        namespace_key: &str,
        rule_id: &str,
    ) -> Option<Vec<flipt::EvaluationDistribution>> {
        let namespace = self.namespace.get(namespace_key)?;

        let evaluation_distributions = namespace.eval_distributions.get(rule_id)?;

        Some(evaluation_distributions.to_vec())
    }

    fn get_evaluation_rollouts(
        &self,
        namespace_key: &str,
        flag_key: &str,
    ) -> Option<Vec<flipt::EvaluationRollout>> {
        let namespace = self.namespace.get(namespace_key)?;

        let eval_rollouts = namespace.eval_rollouts.get(flag_key)?;

        Some(eval_rollouts.to_vec())
    }
}

#[cfg(test)]
mod tests {
    use super::{Snapshot, Store};
    use crate::models::common;
    use crate::models::flipt;
    use crate::store::parsers::TestParser;

    #[test]
    fn test_snapshot() {
        let tp = TestParser::new(vec!["default".into()]);

        let snapshot = match Snapshot::build(&tp) {
            Ok(s) => s,
            Err(e) => panic!("{}", e),
        };

        let flag_variant = snapshot
            .get_flag("default", "flag1")
            .expect("flag1 should exist");

        assert_eq!(flag_variant.key, "flag1");
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
            flipt::EvaluationSegment {
                segment_key: "segment1".into(),
                match_type: common::SegmentMatchType::Any,
                constraints: vec![flipt::EvaluationConstraint {
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
            flipt::EvaluationSegment {
                segment_key: "segment1".into(),
                match_type: common::SegmentMatchType::Any,
                constraints: vec![flipt::EvaluationConstraint {
                    r#type: common::ConstraintComparisonType::String,
                    property: "fizz".into(),
                    operator: "eq".into(),
                    value: "buzz".into(),
                }],
            }
        );
    }
}
