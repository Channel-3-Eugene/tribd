# tribd

Receiver:

1. Receive a packet.
2. Strip PCR is any.
3. Forward to sender.

FIFO Buffer:

1. Accept incoming from receiver.
2. Wait for token from token bucket controller.
2. Send to sender.

Sender:

1. Receive from buffer.
2. Wait for cadence / PLL / huh?
3. Update payload PCR in PES and PSI stream headers.
3. Stream out.