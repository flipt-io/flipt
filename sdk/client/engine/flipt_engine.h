#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * # Safety
 *
 * This function will initialize an Engine and return a pointer back to the caller.
 */
void *initialize_engine(const char *const *namespaces);

/**
 * # Safety
 *
 * This function will take in a pointer to the engine and return a variant evaluation response.
 */
const char *variant(void *engine_ptr, const char *evaluation_request);

/**
 * # Safety
 *
 * This function will take in a pointer to the engine and return a boolean evaluation response.
 */
const char *boolean(void *engine_ptr, const char *evaluation_request);

/**
 * # Safety
 *
 * This function will free the memory occupied by the engine.
 */
void destroy_engine(void *engine_ptr);
