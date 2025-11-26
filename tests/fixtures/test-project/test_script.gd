extends Node

func _ready() -> void:
	print("Test script starting...")

	var result: int = calculate_sum(5, 10)
	print("Result: ", result)

	test_loop()

	print("Test script finished setup!")

var _process_counter: int = 0
var _accumulated_time: float = 0.0

func _process(delta: float) -> void:
	_process_counter += 1
	_accumulated_time += delta

	if _process_counter % 60 == 0:
		print("Process running... Count: ", _process_counter, " Time: ", _accumulated_time)
		temp_var = _process_counter * 2 # Breakpoint here to test local vars

var temp_var: int = 0 # Declared as a member variable

func calculate_sum(a: int, b: int) -> int:
	var sum: int = a + b  # Good place for breakpoint (line 13)
	return sum

func test_loop() -> void:
	for i in range(3):
		print("Loop iteration: ", i)  # Another breakpoint location (line 18)
		var squared: int = i * i
		print("  Squared: ", squared)
