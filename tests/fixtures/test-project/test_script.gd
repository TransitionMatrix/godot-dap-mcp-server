extends Node

func _ready() -> void:
	print("Test script starting...")

	var result: int = calculate_sum(5, 10)
	print("Result: ", result)

	test_loop()

	print("Test script finished!")


func calculate_sum(a: int, b: int) -> int:
	var sum: int = a + b  # Good place for breakpoint (line 13)
	return sum

func test_loop() -> void:
	for i in range(3):
		print("Loop iteration: ", i)  # Another breakpoint location (line 18)
		var squared: int = i * i
		print("  Squared: ", squared)
