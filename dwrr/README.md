# DWRR (Deficit Weighted Round Robin) Scheduler

The DWRR Scheduler is an implementation of the Deficit Weighted Round Robin algorithm in Go. It supports scheduling of any type (`T`) and is designed to handle dynamic queue management and efficient scheduling based on defined quantums and a maximum take limit.

## Overview

Deficit Weighted Round Robin (DWRR) is a scheduling algorithm often used in packet-switched networks to ensure that each queue (or traffic class) receives a fair amount of service relative to its assigned weight. Unlike the strict Round Robin or the Weighted Round Robin, DWRR allows for more flexibility by using a deficit counter, which carries over any unused quantum from one round to the next, allowing queues with larger packets or items a fair chance to transmit over multiple rounds.

## Features

- **Generic Implementation**: Works with any data type, allowing scheduling of diverse item types beyond just packets.
- **Dynamic Queue Management**: Easily add or remove queues at runtime.
- **Quantum Flexibility**: Each queue has a quantum that defines its weight relative to others.
- **Max Take Limit**: Ensures that no queue monopolizes the service by limiting the maximum number of items processed in a single operation cycle.

## Modifications from Standard DWRR

This implementation introduces several key modifications that enhance performance and usability:

- **Generics**: By utilizing Go's generics, this DWRR scheduler can handle any type of item, not just network packets, making it versatile for various applications.
- **Dynamic Quantum Adjustment**: The quantums are adjusted dynamically based on the current queue length, ensuring that the scheduler remains fair and responsive to real-time changes in queue sizes.
- **Deficits**: Setting the quantums from the length of the queue immediately after taking items serves to dynamically adjust the quantums and also factor the deficits while setting the quantum.
- **Max Take Control**: By introducing a `maxTake` limit, the scheduler prevents any single queue from dominating the processing time during a cycle, which is crucial for maintaining fairness in a system with highly variable queue lengths.

## Performance Advantages

This DWRR implementation offers several performance advantages that differentiate it from traditional implementations:

- **Simplified Quantum Management**: Unlike standard DWRR implementations that may require dynamic adjustments to quantum values based on ongoing traffic analysis, this implementation initializes quantum values once and adjusts only through direct operations on the queues. This simplification reduces the computational overhead significantly.

- **No Separate Deficit Term**: Traditional DWRR algorithms often maintain a deficit counter for each queue to track the difference between the quantum and the actual amount processed. In this implementation, the quantum itself is adjusted directly, eliminating the need for separate deficit tracking. This reduces the complexity of the code and the amount of state that must be managed, leading to faster operations.

- **Whole Packet Counting**: By focusing on whole packets, rather than bytes or other measurements, this implementation avoids the computational overhead associated with fragmenting or combining data to meet precise quantum sizes. This is particularly beneficial in environments where data naturally fits into discrete packets, making the system both simpler and faster.

## Usage

To use this scheduler, create an instance of the DWRR with the desired number of queues and maximum take limit, then dynamically manage queues and items as required:

```go
package main

import (
    "fmt"
    "dwrr"
)

func main() {
    scheduler := dwrr.NewDWRR[string](5, 2)  // Initialize with 5 queues and maxTake of 2
    scheduler.Enqueue(0, []string{"item1", "item2", "item3"})
    output := scheduler.Do()
    fmt.Println(output)  // Outputs the items processed in the current cycle
}
