use libc::c_void;
use snafu::{prelude::*, Whatever};
use std::ffi::CStr;
use std::os::raw::c_char;
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
        let update_interval = std::env::var("FLIPT_UPDATE_INTERVAL")
            .unwrap_or("120".into())
            .parse::<u64>()
            .unwrap_or(120);

        std::thread::spawn(move || loop {
            std::thread::sleep(std::time::Duration::from_secs(update_interval));
            let mut lock = evaluator.lock().unwrap();
            lock.replace_snapshot();
        });
    }

    pub fn variant(
        &self,
        evaluation_request: &evaluator::EvaluationRequest,
    ) -> evaluator::VariantEvaluationResponse {
        let binding = self.evaluator.clone();
        let lock = binding.lock().unwrap();

        let variant_eval_response = lock.variant(evaluation_request);

        variant_eval_response.unwrap()
    }

    pub fn boolean(
        &self,
        evaluation_request: &evaluator::EvaluationRequest,
    ) -> evaluator::BooleanEvaluationResponse {
        let binding = self.evaluator.clone();
        let lock = binding.lock().unwrap();

        let boolean_eval_response = lock.boolean(evaluation_request);

        boolean_eval_response.unwrap()
    }
}

/// # Safety
///
/// This function should not be called unless an Engine is initiated. It provides a helper
/// utility to retrieve an Engine instance for evaluation use.
unsafe fn get_engine<'a>(engine_ptr: *mut c_void) -> Result<&'a mut Engine, Whatever> {
    if engine_ptr.is_null() {
        whatever!("null pointer engine error");
    } else {
        Ok(unsafe { &mut *(engine_ptr as *mut Engine) })
    }
}

/// # Safety
///
/// This function will initialize an Engine and return a pointer back to the caller.
#[no_mangle]
pub unsafe extern "C" fn initialize_engine(namespaces: *const *const c_char) -> *mut c_void {
    let mut index = 0;
    let mut namespaces_vec = Vec::new();

    while !(*namespaces.offset(index)).is_null() {
        let c_str = CStr::from_ptr(*namespaces.offset(index));
        if let Ok(rust_str) = c_str.to_str() {
            namespaces_vec.push(rust_str.to_string());
        }

        index += 1;
    }

    Box::into_raw(Box::new(Engine::new(namespaces_vec))) as *mut c_void
}

/// # Safety
///
/// This function will free the memory occupied by the engine.
#[no_mangle]
pub unsafe extern "C" fn destroy_engine(engine_ptr: *mut c_void) {
    if engine_ptr.is_null() {
        return;
    }

    drop(Box::from_raw(engine_ptr as *mut Engine));
}
