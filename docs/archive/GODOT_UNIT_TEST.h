/**************************************************************************/
/*  test_debug_adapter_dictionary.h                                       */
/**************************************************************************/
/*                         This file is part of:                          */
/*                             GODOT ENGINE                               */
/*                        https://godotengine.org                         */
/**************************************************************************/
/* Copyright (c) 2014-present Godot Engine contributors (see AUTHORS.md). */
/* Copyright (c) 2007-2014 Juan Linietsky, Ariel Manzur.                  */
/*                                                                        */
/* Permission is hereby granted, free of charge, to any person obtaining  */
/* a copy of this software and associated documentation files (the        */
/* "Software"), to deal in the Software without restriction, including    */
/* without limitation the rights to use, copy, modify, merge, publish,    */
/* distribute, sublicense, and/or sell copies of the Software, and to     */
/* permit persons to whom the Software is furnished to do so, subject to  */
/* the following conditions:                                              */
/*                                                                        */
/* The above copyright notice and this permission notice shall be         */
/* included in all copies or substantial portions of the Software.        */
/*                                                                        */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,        */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF     */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,   */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE      */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                 */
/**************************************************************************/

#pragma once

#include "core/variant/dictionary.h"
#include "tests/test_macros.h"

namespace TestDebugAdapterDictionary {

// These tests demonstrate the Dictionary safety pattern used in the DAP implementation.
// The DAP protocol allows most fields to be optional, but unsafe Dictionary access
// patterns (using operator[]) cause crashes when clients omit optional fields.
//
// Related: Issue #XXXXX - DAP Dictionary access crashes

TEST_CASE("[DebugAdapter] Dictionary.get() safely handles missing keys") {
	// Simulate a DAP request with missing optional fields
	Dictionary request;
	request["type"] = "request";
	request["command"] = "launch";
	// Deliberately omit "arguments" field (optional per DAP spec)

	// UNSAFE: This would crash with "Dictionary::operator[] used when there was no value"
	// Dictionary args = request["arguments"];

	// SAFE: This returns an empty Dictionary without crashing
	Dictionary args = request.get("arguments", Dictionary());
	CHECK(args.is_empty());
}

TEST_CASE("[DebugAdapter] Dictionary.get() with defaults for nested fields") {
	// Simulate a DAP request with missing nested optional fields
	Dictionary request;
	request["type"] = "request";
	request["command"] = "initialize";

	Dictionary args;
	args["clientID"] = "test-client";
	// Deliberately omit "clientName" (optional per DAP spec)
	request["arguments"] = args;

	// Safe access pattern
	Dictionary arguments = request.get("arguments", Dictionary());
	String client_id = arguments.get("clientID", "");
	String client_name = arguments.get("clientName", "Unknown");

	CHECK(client_id == "test-client");
	CHECK(client_name == "Unknown"); // Default value used
}

TEST_CASE("[DebugAdapter] Dictionary.get() preserves existing values") {
	// Verify that .get() correctly returns existing values
	Dictionary request;
	request["type"] = "request";
	request["seq"] = 42;
	request["command"] = "setBreakpoints";

	Dictionary args;
	args["source"] = "/path/to/file.gd";
	request["arguments"] = args;

	// Safe access should retrieve the actual values
	Dictionary arguments = request.get("arguments", Dictionary());
	String source = arguments.get("source", "");

	CHECK(request.get("seq", 0) == 42);
	CHECK(request.get("command", "") == "setBreakpoints");
	CHECK(source == "/path/to/file.gd");
}

TEST_CASE("[DebugAdapter] Dictionary.has() can check before access") {
	// Alternative pattern: check existence before accessing
	Dictionary request;
	request["type"] = "request";
	request["command"] = "launch";

	// Pattern 1: Check with has() then use get()
	if (request.has("arguments")) {
		Dictionary args = request.get("arguments", Dictionary());
		// Process args
	}

	// Pattern 2: Use get() with default (simpler, preferred)
	Dictionary args = request.get("arguments", Dictionary());
	CHECK(args.is_empty()); // No arguments provided
}

TEST_CASE("[DebugAdapter] Type coercion with safe Dictionary access") {
	// DAP requests can have various types - ensure safe handling
	Dictionary request;
	request["type"] = "request";
	request["seq"] = 123;

	// Safe integer extraction with default
	int seq = request.get("seq", 0);
	CHECK(seq == 123);

	// Missing field returns default
	int missing = request.get("missing_field", -1);
	CHECK(missing == -1);

	// Type mismatch handling
	String seq_as_string = request.get("seq", ""); // Type mismatch
	// Godot will attempt conversion, but default is safe fallback
}

} // namespace TestDebugAdapterDictionary
