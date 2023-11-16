use libc::c_void;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use snafu::{prelude::*, Whatever};
use std::collections::HashMap;
use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::sync::{Arc, Mutex};

pub mod evaluator;
mod models;
mod store;

#[derive(Deserialize)]
struct EvaluationReq {
    namespace_key: String,
    flag_key: String,
    entity_id: String,
    context: String,
}

pub struct Engine {
    pub evaluator: Arc<Mutex<evaluator::Evaluator>>,
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
    ) -> Result<evaluator::VariantEvaluationResponse, Whatever> {
        let binding = self.evaluator.clone();
        let lock = binding.lock().unwrap();

        lock.variant(evaluation_request)
    }

    pub fn boolean(
        &self,
        evaluation_request: &evaluator::EvaluationRequest,
    ) -> Result<evaluator::BooleanEvaluationResponse, Whatever> {
        let binding = self.evaluator.clone();
        let lock = binding.lock().unwrap();

        lock.boolean(evaluation_request)
    }
}

#[derive(Serialize)]
struct FFIResponse<T>
where
    T: Serialize,
{
    status: Status,
    result: Option<T>,
    error_message: Option<String>,
}

#[derive(Serialize)]
enum Status {
    #[serde(rename = "success")]
    Success,
    #[serde(rename = "failure")]
    Failure,
}

impl<T> From<Result<T, Whatever>> for FFIResponse<T>
where
    T: Serialize,
{
    fn from(value: Result<T, Whatever>) -> Self {
        match value {
            Ok(result) => FFIResponse {
                status: Status::Success,
                result: Some(result),
                error_message: None,
            },
            Err(e) => FFIResponse {
                status: Status::Failure,
                result: None,
                error_message: Some(e.to_string()),
            },
        }
    }
}

fn result_to_json_ptr<T: Serialize>(result: Result<T, Whatever>) -> *mut c_char {
    let ffi_response: FFIResponse<T> = result.into();
    let json_string = serde_json::to_string(&ffi_response).unwrap();
    CString::new(json_string).unwrap().into_raw()
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
/// This function will take in a pointer to the engine and return a variant evaluation response.
#[no_mangle]
pub unsafe extern "C" fn variant(
    engine_ptr: *mut c_void,
    evaluation_request: *const c_char,
) -> *const c_char {
    let e = get_engine(engine_ptr).unwrap();
    let e_req = get_evaluation_request(evaluation_request);

    result_to_json_ptr(e.variant(&e_req))
}

/// # Safety
///
/// This function will take in a pointer to the engine and return a boolean evaluation response.
#[no_mangle]
pub unsafe extern "C" fn boolean(
    engine_ptr: *mut c_void,
    evaluation_request: *const c_char,
) -> *const c_char {
    let e = get_engine(engine_ptr).unwrap();
    let e_req = get_evaluation_request(evaluation_request);

    result_to_json_ptr(e.boolean(&e_req))
}

unsafe fn get_evaluation_request(
    evaluation_request: *const c_char,
) -> evaluator::EvaluationRequest {
    let evaluation_request_bytes = CStr::from_ptr(evaluation_request).to_bytes();
    let bytes_str_repr = std::str::from_utf8(evaluation_request_bytes).unwrap();

    let client_eval_request: EvaluationReq = serde_json::from_str(bytes_str_repr).unwrap();

    let parsed_context: serde_json::Value =
        serde_json::from_str(&client_eval_request.context).unwrap();

    let mut context_map: HashMap<String, String> = HashMap::new();
    if let serde_json::Value::Object(map) = parsed_context {
        for (key, value) in map {
            if let Value::String(val) = value {
                context_map.insert(key, val);
            }
        }
    }

    evaluator::EvaluationRequest {
        namespace_key: client_eval_request.namespace_key,
        flag_key: client_eval_request.flag_key,
        entity_id: client_eval_request.entity_id,
        context: context_map,
    }
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
