use std::sync::{Arc, Mutex};

mod common;
pub mod evaluator;
mod flipt;
mod store;

pub struct Engine {
    pub evaluator: Arc<Mutex<evaluator::Evaluator<store::parsers::FliptParser>>>,
}

impl Engine {
    pub fn new(namespaces: Vec<String>) -> Self {
        let evaluation_engine = evaluator::Evaluator::new(namespaces);

        let mut evaluator = Self {
            evaluator: Arc::new(Mutex::new(evaluation_engine.unwrap())),
        };

        evaluator.update();

        evaluator
    }

    fn update(&mut self) {
        let evaluator = self.evaluator.clone();
        std::thread::spawn(move || loop {
            std::thread::sleep(std::time::Duration::from_millis(2000));
            let mut lock = evaluator.lock().unwrap();
            lock.replace_snapshot();
        });
    }

    pub fn variant(
        &self,
        evaluation_request: &evaluator::EvaluationRequest,
    ) -> evaluator::VariantEvaluationResponse {
        let binding = self.evaluator.clone();
        let mut lock = binding.lock().unwrap();

        let variant_eval_response = lock.variant(evaluation_request);

        variant_eval_response.unwrap()
    }
}
