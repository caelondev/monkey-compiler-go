import sys
import time

sys.setrecursionlimit(50000)  # Default is ~1000

def ack(m, n):
    if m == 0:
        return n + 1
    if n == 0:
        return ack(m - 1, 1)
    return ack(m - 1, ack(m, n - 1))

start = time.perf_counter()
result = ack(3, 10)
end = time.perf_counter()

print(f"Result: {result}")
print(f"Time: {(end - start) * 1000:.4f} ms")
