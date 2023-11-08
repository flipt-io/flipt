// cargo run --example evaluation

use engine::{self, evaluator};
use std::collections::HashMap;

fn main() {
    let eng = engine::Engine::new(vec!["default".into()]);
    let mut context: HashMap<String, String> = HashMap::new();
    context.insert("fizz".into(), "buzz".into());

    let thread = std::thread::spawn(move || loop {
        std::thread::sleep(std::time::Duration::from_millis(5000));
        let variant = eng.variant(&evaluator::EvaluationRequest {
            namespace_key: "default".into(),
            flag_key: "flag1".into(),
            entity_id: "entity".into(),
            context: context.clone(),
        });

        println!("variant key {}", variant.unwrap().variant_key);
    });

    thread.join().expect("current thread panicked");
}
