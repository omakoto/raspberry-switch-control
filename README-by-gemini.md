# raspberry-switch-control

[Made by Gemini, not reviewed by human yet.]

Emulate a Nintendo Switch USB controller (Pro Controller) using a Raspberry Pi's USB gadget mode. This project allows you to control a Nintendo Switch either interactively via text commands on the command line, through automated scripts, or by forwarding input from a physical controller (Nintendo Pro Controller, Xbox One Controller, PS4 Controller) connected to a host PC.

Credit: This project is based on and heavily relies on [mzyy94/nscon](https://github.com/mzyy94/nscon).

---

## Features

- **USB Gadget Emulation**: Configures a Raspberry Pi to present itself to the Nintendo Switch as an official Nintendo Switch Pro Controller.
- **Direct Interactive Backend (`nsbackend`)**: Accepts real-time commands from stdin or a Unix FIFO pipeline to trigger button presses, D-pad events, or analog stick adjustments.
- **Precision Timed Sequences**: Support for duration prefixes on commands (e.g., press a button for exactly 50ms or hold it for `N` seconds).
- **Host PC Frontend (`nsfrontend`)**: Reads input from a physical controller connected to a PC and pipes those actions directly to the Raspberry Pi over SSH or a local FIFO.
- **Multiple Controller Support**: Mappings pre-configured for Nintendo Switch Pro Controllers, Xbox One Controllers, and Sony DualShock 4 Controllers.

---

## Hardware Requirements

- **Raspberry Pi**: Tested and confirmed working with **Raspberry Pi Zero W**, **Zero 2 W**, and **Pi 4**.

  - *Note on Raspberry Pi 5*:
    > [!IMPORTANT]
    > **This part is totally unverified yet**

    The Raspberry Pi 5 can support USB gadget mode, but its setup differs from previous models because of its updated I/O architecture and Raspberry Pi OS's transition to `NetworkManager` (replacing `dhcpcd`). 
    - To configure peripheral mode on Pi 5, use `dtoverlay=dwc2,dr_mode=peripheral` in `/boot/firmware/config.txt` and append `modules-load=dwc2` in `/boot/firmware/cmdline.txt`.
    - Additionally, because of the Pi 5's higher power requirements and its PMIC power negotiation over the USB-C port, you may experience stability/negotiation issues when connecting it directly to the Switch. Powering the Pi 5 externally (e.g., via the 5V GPIO header) or using a suitable powered hub is strongly recommended.
    - For a detailed setup walkthrough, refer to [Ben Hardill's Guide on Raspberry Pi 5 USB Gadget](https://www.hardill.me.uk/wp/2023/12/22/raspberry-pi-5-usb-gadget/).
- **Connection Cables**:
  - For **Pi Zero / Zero W / Zero 2 W**: A micro-USB data sync cable connected to the USB OTG port.
  - For **Pi 4**: A USB-C cable connected to the USB-C port.
- **Power Configuration (Recommended)**:
  - To prevent the Raspberry Pi from rebooting when the Nintendo Switch goes to sleep, is docked, or is disconnected, it is highly recommended to connect the Pi to a **powered USB hub**, and then connect that hub to the Switch. Make sure the Pi is drawing sufficient power.

---

## Project Structure & Scripts

- [00install.bash](./00install.bash): Installs the `nsbackend` and `nsfrontend` binaries to your Go bin path (`$GOPATH/bin` or `$HOME/go/bin`).
- [1-run-backend](./1-run-backend): Wrapper to run the backend command locally using `go run`.
- [2-run-frontend](./2-run-frontend): Wrapper to run the frontend command locally using `go run`.
- [presubmit.sh](./presubmit.sh): Runs code formatting (`gofmt`), static analysis (`go vet`), and tests.
- [nscontroller/cmd/nsbackend](./nscontroller/cmd/nsbackend/): Main backend binary that runs on the Raspberry Pi and writes reports directly to `/dev/hidg0`.
- [nscontroller/cmd/nsfrontend](./nscontroller/cmd/nsfrontend/): Main frontend binary that runs on the host PC and reads from physical joystick inputs (`/dev/input/jsX`).
- [nscontroller/scripts/switch-controller-gadget](./nscontroller/scripts/switch-controller-gadget): The shell script that configures `/sys/kernel/config/usb_gadget/` to enable Switch Controller emulation.

---

## Setup Instructions

### 1. Configure Raspberry Pi for USB Gadget Mode

This enables the Raspberry Pi to function as a USB OTG device. Tested on **Ubuntu 24** (on Pi 4) and **Raspberry Pi OS** (on Zero W 2).

1. Edit `/boot/firmware/config.txt` (or `/boot/config.txt` on older OS releases) to enable the OTG driver:
   ```bash
   echo "dtoverlay=dwc2" | sudo tee -a /boot/firmware/config.txt
   ```
2. Enable the required modules in `/etc/modules`:
   ```bash
   echo "dwc2" | sudo tee -a /etc/modules
   echo "libcomposite" | sudo tee -a /etc/modules
   ```
3. Reboot the Raspberry Pi:
   ```bash
   sudo reboot
   ```

### 2. Install Binaries

#### On the Raspberry Pi:
Install the dependencies and build the backend binary:
```bash
sudo apt install -y golang xxd git

# Option A: Install directly via Go command
go install -v github.com/omakoto/raspberry-switch-control/nscontroller/cmd/...@latest

# Option B: Clone this repo and run:
./00install.bash
```

#### On the Host PC (if using joystick forwarding):
Install Go and the frontend binary:
```bash
sudo apt install -y golang

# Option A: Install via Go command
go install -v github.com/omakoto/raspberry-switch-control/nscontroller/cmd/...@latest

# Option B: Run inside the cloned directory:
./00install.bash
```

### 3. Initialize the USB Gadget

You must run the gadget script on the Raspberry Pi as root to register `/dev/hidg0`.

1. Find the path of the gadget script dynamically:
   ```bash
   nsbackend usb-init-script-path
   ```
2. Run the script:
   ```bash
   sudo bash "$($(go env GOPATH)/bin/nsbackend usb-init-script-path)"
   ```
3. (Optional) Run the script automatically at boot time by adding it to root's crontab (`sudo crontab -e`):
   ```crontab
   @reboot bash -c ". $(/home/pi/go/bin/nsbackend usb-init-script-path)"
   ```
   *(Adjust paths depending on your user account name and where Go installs binaries).*

### 4. Connect to the Switch

- **Raspberry Pi 4**: Connect the USB-C port of the Pi to the Switch (optionally through a powered USB hub).
- **Raspberry Pi Zero**: Connect the micro-USB OTG/Data port (not the power-only port) of the Pi to the Switch.

---

## Running the Emulation

There are three primary ways to use the controller emulation.

### Mode A: Direct Interactive Console / Scripting (No Joystick Needed)

You can send controller commands directly to `nsbackend` running on the Pi:
```bash
# Starts nsbackend interactively reading commands from stdin
nsbackend -f /dev/hidg0
```
Type commands (e.g. `a` to press A, `ly 1` to push left stick up) and press Enter.

You can also pipe scripts to automate tasks:
```bash
cat << 'EOF' | nsbackend
0.5 a
0.5 b
0.5 lx 1
0.5 lx 0
EOF
```

### Mode B: Direct Joystick Forwarding from PC

Map a physical controller plugged into your host PC to the Raspberry Pi over SSH:
```bash
# Run on Host PC (assuming $PI_ADDRESS is your Pi's hostname/IP)
nsfrontend -j /dev/input/js0 -o >(ssh pi@$PI_ADDRESS go/bin/nsbackend)
```
Press `[Enter]` in the console to finish and disconnect.

### Mode C: Background Daemon with FIFO Pipeline (Advanced)

To avoid launching ssh connections repeatedly, run the backend on the Pi as a persistent daemon:

1. **On the Raspberry Pi**:
   Add this to `root`'s crontab (`sudo crontab -e`) to start the daemon at boot:
   ```crontab
   @reboot bash -c ". $(/home/pi/go/bin/nsbackend usb-init-script-path); /home/pi/go/bin/nsbackend -x"
   ```
   *Alternatively, run it manually in the background*:
   ```bash
   nsbackend -x
   ```
   *(Note: `-x` or `--daemon` runs the backend in the background and listens for inputs at `/tmp/nsbackend.fifo` by default).*

2. **On the Host PC**:
   Route your joystick frontend output into the remote daemon's FIFO:
   ```bash
   nsfrontend -j /dev/input/js0 -o >(ssh pi@$PI_ADDRESS 'cat > /tmp/nsbackend.fifo')
   ```

---

## Command Protocol & Stdin Syntax

`nsbackend` accepts commands via stdin/FIFO where each line is evaluated.

### Format
```
[[duration] ]command [value]
```

- **Default Action (Autorelease)**: If no value is specified (e.g. `a`), the button is pressed and automatically released after a default window (50ms).
- **Explicit Toggle**: If a value is provided (e.g. `a 1` or `a 0`), the state is set and held.
- **Duration Prefix**: Specifying a leading number (seconds) controls timing:
  - For autorelease commands (e.g., `0.5 a`): Presses A, releases it after 50ms, and sleeps/blocks the coordinator queue for the remaining 450ms.
  - For explicit commands (e.g., `0.5 a 1` followed by `0.5 a 0`): Sets A to 1, waits 500ms, sets A to 0, and waits another 500ms.

### Button Codes
| Code | Switch Controller Button |
|---|---|
| `a` | A Button |
| `b` | B Button |
| `x` | X Button |
| `y` | Y Button |
| `h` | Home Button |
| `c` | Capture Button |
| `-` or `m` | Minus Button |
| `+` or `p` | Plus Button |
| `l1` | L Button |
| `l2` | ZL Button |
| `r1` | R Button |
| `r2` | ZR Button |
| `lp` | Left Stick Press |
| `rp` | Right Stick Press |
| `pu` / `pd` / `pl` / `pr` | D-pad Up / Down / Left / Right |
| `pul` / `pur` / `pdl` / `pdr` | D-pad Diagonals (Up-Left, Up-Right, Down-Left, Down-Right) |

### Stick & Analog Codes
- Analog sticks take a float value from `-1.0` to `1.0` (center is `0.0`).
- D-pad analog alternatives (`px`/`py`) trigger digital buttons based on thresholds.

| Code | Input Element | Behavior |
|---|---|---|
| `lx` | Left Stick X | `-1.0` (Left) to `1.0` (Right) |
| `ly` | Left Stick Y | `1.0` (Up) to `-1.0` (Down) |
| `rx` | Right Stick X | `-1.0` (Left) to `1.0` (Right) |
| `ry` | Right Stick Y | `1.0` (Up) to `-1.0` (Down) |
| `px` | D-pad X | `< -1.0` presses D-pad Left, `> 1.0` presses D-pad Right |
| `py` | D-pad Y | `< -1.0` presses D-pad Up, `> 1.0` presses D-pad Down |

#### Stick Shortcuts
You can also use quick digital-style triggers for analog sticks. Sending these values sets the stick fully in that direction, and they support auto-release:
- **Left Stick**: `lu` (Up), `ld` (Down), `ll` (Left), `lr` (Right), `lul` (Up-Left), `lur` (Up-Right), `ldl` (Down-Left), `ldr` (Down-Right).
- **Right Stick**: `ru` (Up), `rd` (Down), `rl` (Left), `rr` (Right), `rul` (Up-Left), `rur` (Up-Right), `rdl` (Down-Left), `rdr` (Down-Right).

---

## TODOs & Missing Features

- **Autofire support**: An internal autofire loop engine is implemented in the codebase (`autofire.go`), but user-facing controls to dynamically enable, disable, and configure it are not yet fully wired.
- **Macro support**: Automated macro definitions and recording.

---

## References

1. **Nintendo Switch Pro Controller USB Gadget**: https://mzyy94.com/blog/2020/03/20/nintendo-switch-pro-controller-usb-gadget/
2. **USB Gadget API for Linux**: https://www.kernel.org/doc/html/v4.13/driver-api/usb/gadget.html
3. **Raspberry Pi Joystick project**: https://github.com/milador/RaspberryPi-Joystick
4. **LUFA Library USB Descriptor Documentation**: http://www.fourwalledcubicle.com/files/LUFA/Doc/120219/html/group___group___std_descriptors.html
