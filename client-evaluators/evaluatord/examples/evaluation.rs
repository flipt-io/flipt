// cargo run --example evaluation

use evaluatord::evaluator;
use std::collections::HashMap;

#[tokio::main]
async fn main() {
    let evaluation_engine = evaluator::Evaluator::new(vec!["default".into()]).unwrap();

    let mut context: HashMap<String, String> = HashMap::new();
    context.insert("fizz".into(), "buzz".into());

    let variant_result = evaluation_engine
        .variant(&evaluator::EvaluationRequest {
            namespace_key: "default".into(),
            flag_key: "flag1".into(),
            entity_id: "newentityid".into(),
            context: context.clone(),
        })
        .unwrap();

    let boolean_result = evaluation_engine
        .boolean(&evaluator::EvaluationRequest {
            namespace_key: "default".into(),
            flag_key: "flag1".into(),
            entity_id: "newentityid".into(),
            context: context,
        })
        .unwrap();

    println!("{}", variant_result.variant_key);
    println!("{}", boolean_result.enabled);
}
