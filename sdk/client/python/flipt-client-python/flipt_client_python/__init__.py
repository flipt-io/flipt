import ctypes
import json
import os

from .models import (
    BooleanEvaluationResponse,
    EvaluationRequest,
    VariantEvaluationResponse,
)


class FliptEvaluationClient:
    def __init__(self, namespaces: list[str]):
        engine_library_path = os.environ.get("ENGINE_LIB_PATH")
        if engine_library_path is None:
            raise Exception("ENGINE_LIB_PATH not set")

        self.ffi_core = ctypes.CDLL(engine_library_path)

        self.ffi_core.initialize_engine.restype = ctypes.c_void_p
        self.ffi_core.destroy_engine.argtypes = [ctypes.c_void_p]

        self.ffi_core.variant.argtypes = [ctypes.c_void_p, ctypes.c_char_p]
        self.ffi_core.variant.restype = ctypes.c_char_p

        self.ffi_core.boolean.argtypes = [ctypes.c_void_p, ctypes.c_char_p]
        self.ffi_core.boolean.restype = ctypes.c_char_p

        ns = (ctypes.c_char_p * len(namespaces))()
        ns[:] = [s.encode("utf-8") for s in namespaces]

        self.engine = self.ffi_core.initialize_engine(ns)

    def __del__(self):
        if hasattr(self, "engine") and self.engine is not None:
            self.destroy_engine(self.engine)

    def variant(
        self, namespace_key: str, flag_key: str, entity_id: str, context: dict
    ) -> VariantEvaluationResponse:
        response = self.ffi_core.variant(
            self.engine,
            serialize_evaluation_request(namespace_key, flag_key, entity_id, context),
        )

        bytes_returned = ctypes.c_char_p(response).value

        variant_evaluation_response = VariantEvaluationResponse.parse_raw(
            bytes_returned
        )

        return variant_evaluation_response

    def boolean(
        self, namespace_key: str, flag_key: str, entity_id: str, context: dict
    ) -> BooleanEvaluationResponse:
        response = self.ffi_core.boolean(
            self.engine,
            serialize_evaluation_request(namespace_key, flag_key, entity_id, context),
        )

        bytes_returned = ctypes.c_char_p(response).value

        boolean_evaluation_response = BooleanEvaluationResponse.parse_raw(
            bytes_returned
        )

        return boolean_evaluation_response


def serialize_evaluation_request(
    namespace_key: str, flag_key: str, entity_id: str, context: dict
) -> str:
    evaluation_request = EvaluationRequest(
        namespace_key=namespace_key,
        flag_key=flag_key,
        entity_id=entity_id,
        context=json.dumps(context),
    )

    return evaluation_request.json().encode("utf-8")
