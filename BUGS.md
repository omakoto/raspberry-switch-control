# Identified Bugs in Raspberry Switch Control

This document lists identified bugs, data races, and potential panic scenarios in the codebase.

**These are raw bug reports made by Gemini 3.5 flash medium. Not all of them may be valid**

## 1. Out-of-bounds Panic on Duration-Only Commands in `nsbackend.go`
- **Location**: `nscontroller/cmd/nsbackend/nsbackend.go`
- **Description**: If a user inputs a duration-only command like `"1.5"`, `strings.Fields(s)` returns a slice of length 1. Because `'0' <= arr[0][0] && arr[0][0] <= '9'` is true, it parses the duration and sets `cmdIndex = 1`. Next, the code attempts to read `command = arr[cmdIndex]`, which evaluates to `arr[1]`. Since `len(arr)` is 1, this causes an immediate index out-of-bounds panic.

## 2. Out-of-bounds Panic in `SetAutofire` for Non-Button Actions
- **Location**: `nscontroller/autofire.go`
- **Description**: `SetAutofire(a Action, ...)` uses the `a` argument directly to index into the `af.states` slice. However, `af.states` is initialized with size `ActionButtonLast` (which is `ActionAxisLX`). If `SetAutofire` is called with `ActionAxisLX` or any axis/synthetic action, it will panic with an index out of bounds because `a >= len(af.states)`.

## 3. Nil-Pointer Panic on Unknown Event Types in `js_linux.go`
- **Location**: `nscontroller/js/js_linux.go`
- **Description**: In the joystick event reader loop, if `oev.EventType &^ jsEventInit` matches neither `jsEventAxis` nor `jsEventButton`, `event.Element` remains `nil`. The code then unconditionally executes `event.Element.Value = event.Value`, which causes a nil-pointer dereference panic and crashes the entire frontend.

## 4. Data Race on `count` in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: `c.count` is incremented concurrently inside the `startCounter` goroutine (`c.count++`) and read inside `startInputReport` (`c.write(0x30, c.count, ...)`) without any synchronization or atomic operations. This is a classic Go data race.

## 5. Swapped Square and Triangle Mappings in `psjoystick.go`
- **Location**: `nscontroller/psjoystick.go`
- **Description**: In the PlayStation (DS4) mapping, button `0x133` (Square) is mapped to `ActionButtonX` (Switch X) and button `0x134` (Triangle) is mapped to `ActionButtonY` (Switch Y). Physically on the controller, Triangle is at the top (which corresponds to X on the Switch Pro Controller) and Square is on the left (which corresponds to Y on the Switch Pro Controller). They are currently swapped.

## 6. Premature Auto-Release Under Overlapping Presses in `streaminput.go`
- **Location**: `nscontroller/streaminput.go`
- **Description**: `StreamInput` doesn't track active button states or cancel/override previous auto-releases. If a user presses a button and then presses it again 10ms later, the first press spawns an auto-release goroutine that will release the button at `+60ms`, prematurely releasing it even though the second press should hold it until `+70ms`.

## 7. Incorrect FIFO Logging Path in `nsbackend.go`
- **Location**: `nscontroller/cmd/nsbackend/nsbackend.go`
- **Description**: When creating a FIFO, the log prints `Creating FIFO at %s...\n` using `input.Name()` (which points to `/dev/stdin` at that point in execution) instead of using the target FIFO path `*fifo`.

## 8. Wrong Warning Arguments on Float Parsing Failures in `nsbackend.go`
- **Location**: `nscontroller/cmd/nsbackend/nsbackend.go`
- **Description**: When parsing duration or stick/button arguments in `parseCommand`, if parsing fails, the warning prints `arr[1]` instead of the actual index that failed to parse (e.g. `arr[0]` or `arr[cmdIndex+1]`). This can also trigger an out-of-bounds panic if `len(arr) == 1`.

## 9. Data Race on `stopInput` Channel in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: The `c.stopInput` channel is closed/reassigned concurrently in `Close()` (main thread) and inside the UART reader goroutine (when receiving command `0x05`) without any mutex protection.

## 10. Endless Error Loop on Joystick Read Failures in `jsinput.go`
- **Location**: `nscontroller/jsinput.go`
- **Description**: If a joystick read error occurs (other than `io.EOF`, e.g., joystick disconnected), `common.Checke(err)` is called. If the application continues after a non-EOF read error, it loops immediately and calls `Read()` again, causing a high-CPU spin on a broken file descriptor.

## 11. L2 and R2 Triggers Non-Functional on PS4 Controller in `psjoystick.go`
- **Location**: `nscontroller/psjoystick.go`
- **Description**: The PlayStation 4 controller exposes its L2 and R2 triggers as axis inputs (`0x02` and `0x05`) in Linux, but `psjoystick.go` only maps button-based inputs (`0x138` and `0x139`). Under standard Linux drivers, trigger presses are not captured as axis inputs, rendering trigger functionality non-functional.

## 12. Nil-Pointer Panic on `c.fp.Write` During Close in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: The `write` method accesses `c.fp.Write()` concurrently with `Close()` setting `c.fp = nil` without synchronization. This can cause a nil-pointer dereference panic when the controller is closed while actively sending input reports.

## 13. Stale/Garbage Data Processing on Short Reads in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: The reader loop does not check the byte count `n` returned by `c.fp.Read(buf)`. If a short HID report or partial packet is read, the code accesses offsets (like `buf[10]` or `buf[12]`) containing stale, garbage data from previous reports, leading to unpredictable controller states.

## 14. Out-of-Bounds Panic on Malformed SPI ROM Reads in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: The SPI ROM read handler slices `data[buf[11]:buf[11]+buf[15]]` using the start offset (`buf[11]`) and read length (`buf[15]`) directly supplied by the host. Because `SPI_ROM_DATA` arrays are short (64 or 176 bytes) and bounds checks are absent, a malformed or malicious read request will panic and crash the daemon.

## 15. Broken File Path Lookup in `usbInitScriptPath` When Deployed
- **Location**: `nscontroller/cmd/nsbackend/subcommands.go`
- **Description**: `usbInitScriptPath` uses `common.GetSourceInfo()` to locate the gadget initialization script. This returns the absolute compilation file path on the host. When deployed on a target machine (like a Raspberry Pi), this path does not exist, causing subcommands like `show-usb-init-script` to fail with a "no such file or directory" error.

## 16. Global Xbox Trigger Initialization States Break on Reconnect
- **Location**: `nscontroller/xboxjoystick.go`
- **Description**: `l2Initialized` and `r2Initialized` are package-level globals. If the controller is disconnected and reconnected, these remain `true`, skipping the trigger initialization logic. Consequently, the new controller starts with a default trigger value of `0.0`, which incorrectly maps as pressed, causing a phantom trigger hold until physically pressed and released.

## 17. Stick Value Overflow Corrupts Packets in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: The stick coordinates `lx, ly, rx, ry` are calculated and cast to `uint16` without clamping. If an analog input drifts or overshoots `[-1..1]`, the values exceed `4095`. Since `packShorts` does not clamp inputs, bits beyond the 12th bit spill over, corrupting the packet contents.

## 18. Duplicate Input Report Goroutines in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: In the UART reader loop, receiving command `0x04` invokes `c.startInputReport()` unconditionally. If the host sends multiple `0x04` requests (common during connection negotiation retries), multiple background report goroutines are spawned, writing reports to `c.fp` concurrently and clogging the bandwidth.

## 19. D-Pad Button Mappings Ignored in Joystick Dispatchers
- **Location**: `nscontroller/nsprojoystick.go`, `nscontroller/xboxjoystick.go`, `nscontroller/psjoystick.go`
- **Description**: The joystick dispatchers ignore D-pad button codes (`0x220` to `0x223`). On setups where the D-pad is mapped by Linux as buttons rather than analog hat axes, the D-pad is completely non-functional.

## 20. Swapped X and Y Buttons in `nsprojoystick.go`
- **Location**: `nscontroller/nsprojoystick.go`
- **Description**: The Switch Pro Controller dispatcher incorrectly swaps the mapping of the physical `X` and `Y` buttons, mapping `0x133` (Y) to `ActionButtonX` and `0x134` (X) to `ActionButtonY`.

## 21. Nil-Pointer Panic on Read During Close in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: When `Close()` is called, the background reader thread unblocks from `Read()` with an error, but the read loop does not terminate. In the next iteration, it attempts to read from `c.fp` (which has been set to `nil`), triggering a nil-pointer dereference panic.

## 22. Slice Index Overflow Wrapping Panic in `nscon.go`
- **Location**: `nscontroller/nscon.go`
- **Description**: In SPI ROM reads, `data[buf[11]:buf[11]+buf[15]]` is sliced. Because `buf[11]` and `buf[15]` are `byte` (uint8) types, adding them can overflow/wrap around `255` (e.g. `200 + 100 = 300` wraps to `44`), causing a start index greater than end index panic (`[200:44]`).

## 23. Inverted Autofire Mode Sends Release When Physically Pressed
- **Location**: `nscontroller/autofire.go`
- **Description**: If `bs.mode` is `AutofireModeInvert`, when the user presses the button (`pressed = true`), the code calls `sendAutofireEventLocked` with `!pressed` (which is `false` / release). This causes the button to stay released when physically held, instead of remaining pressed.

## 24. Fragile Float Comparison for Button Press Detection
- **Location**: `nscontroller/events.go`
- **Description**: The `pressed()` method checks `ev.Value == 1` exactly. Since `Value` is a `float64`, any slightly noisy or modified float input from analog axes/triggers mapped to buttons will fail the exact comparison.

## 25. Frontend Crashes on Partial Read Error During Unplug
- **Location**: `nscontroller/jsinput.go`
- **Description**: Only `io.EOF` is handled as a clean disconnect. If a controller is unplugged, the OS read can fail with `io.ErrUnexpectedEOF` or other read errors, which trigger `common.Checke(err)`, crashing the application with a stack trace instead of exiting cleanly.

## 26. Coordinator Loop Blocked During Synchronous Command Sleeps
- **Location**: `nscontroller/cmd/nsbackend/nsbackend.go`
- **Description**: If a command specifies a duration without auto-release, `sendToController` calls `time.Sleep(dur)` synchronously. Since the coordinator processes commands sequentially on a single goroutine, this blocks all other queued commands, preventing parallel inputs.
