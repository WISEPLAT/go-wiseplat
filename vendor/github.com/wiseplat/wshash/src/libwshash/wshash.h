/*
  This file is part of wshash.

  wshash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  wshash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with wshash.  If not, see <http://www.gnu.org/licenses/>.
*/

/** @file wshash.h
* @date 2015
*/
#pragma once

#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <stddef.h>
#include "compiler.h"

#define WSHASH_REVISION 23
#define WSHASH_DATASET_BYTES_INIT 1073741824U // 2**30
#define WSHASH_DATASET_BYTES_GROWTH 8388608U  // 2**23
#define WSHASH_CACHE_BYTES_INIT 1073741824U // 2**24
#define WSHASH_CACHE_BYTES_GROWTH 131072U  // 2**17
#define WSHASH_EPOCH_LENGTH 30000U
#define WSHASH_MIX_BYTES 128
#define WSHASH_HASH_BYTES 64
#define WSHASH_DATASET_PARENTS 256
#define WSHASH_CACHE_ROUNDS 3
#define WSHASH_ACCESSES 64
#define WSHASH_DAG_MAGIC_NUM_SIZE 8
#define WSHASH_DAG_MAGIC_NUM 0xFEE1DEADBADDCAFE

#ifdef __cplusplus
extern "C" {
#endif

/// Type of a seedhash/blockhash e.t.c.
typedef struct wshash_h256 { uint8_t b[32]; } wshash_h256_t;

// convenience macro to statically initialize an h256_t
// usage:
// wshash_h256_t a = wshash_h256_static_init(1, 2, 3, ... )
// have to provide all 32 values. If you don't provide all the rest
// will simply be unitialized (not guranteed to be 0)
#define wshash_h256_static_init(...)			\
	{ {__VA_ARGS__} }

struct wshash_light;
typedef struct wshash_light* wshash_light_t;
struct wshash_full;
typedef struct wshash_full* wshash_full_t;
typedef int(*wshash_callback_t)(unsigned);

typedef struct wshash_return_value {
	wshash_h256_t result;
	wshash_h256_t mix_hash;
	bool success;
} wshash_return_value_t;

/**
 * Allocate and initialize a new wshash_light handler
 *
 * @param block_number   The block number for which to create the handler
 * @return               Newly allocated wshash_light handler or NULL in case of
 *                       ERRNOMEM or invalid parameters used for @ref wshash_compute_cache_nodes()
 */
wshash_light_t wshash_light_new(uint64_t block_number);
/**
 * Frees a previously allocated wshash_light handler
 * @param light        The light handler to free
 */
void wshash_light_delete(wshash_light_t light);
/**
 * Calculate the light client data
 *
 * @param light          The light client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               an object of wshash_return_value_t holding the return values
 */
wshash_return_value_t wshash_light_compute(
	wshash_light_t light,
	wshash_h256_t const header_hash,
	uint64_t nonce
);

/**
 * Allocate and initialize a new wshash_full handler
 *
 * @param light         The light handler containing the cache.
 * @param callback      A callback function with signature of @ref wshash_callback_t
 *                      It accepts an unsigned with which a progress of DAG calculation
 *                      can be displayed. If all goes well the callback should return 0.
 *                      If a non-zero value is returned then DAG generation will stop.
 *                      Be advised. A progress value of 100 means that DAG creation is
 *                      almost complete and that this function will soon return succesfully.
 *                      It does not mean that the function has already had a succesfull return.
 * @return              Newly allocated wshash_full handler or NULL in case of
 *                      ERRNOMEM or invalid parameters used for @ref wshash_compute_full_data()
 */
wshash_full_t wshash_full_new(wshash_light_t light, wshash_callback_t callback);

/**
 * Frees a previously allocated wshash_full handler
 * @param full    The light handler to free
 */
void wshash_full_delete(wshash_full_t full);
/**
 * Calculate the full client data
 *
 * @param full           The full client handler
 * @param header_hash    The header hash to pack into the mix
 * @param nonce          The nonce to pack into the mix
 * @return               An object of wshash_return_value to hold the return value
 */
wshash_return_value_t wshash_full_compute(
	wshash_full_t full,
	wshash_h256_t const header_hash,
	uint64_t nonce
);
/**
 * Get a pointer to the full DAG data
 */
void const* wshash_full_dag(wshash_full_t full);
/**
 * Get the size of the DAG data
 */
uint64_t wshash_full_dag_size(wshash_full_t full);

/**
 * Calculate the seedhash for a given block number
 */
wshash_h256_t wshash_get_seedhash(uint64_t block_number);

#ifdef __cplusplus
}
#endif
