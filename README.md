# raspberry-switch-control

Emulate Nintendo Switch USB Controller with Raspberry Pi

Credit: This project heavily relies on https://github.com/mzyy94/nscon.

## Hardware requirements
* Raspberry Pi. Tested with Zero W and Pi 4. Not sure if Pi 5 works. (Saw an page saying 5 doesn't support the usb "gadget" mode?)

## Set up Raspberry Pi

Tested on Ubuntu 24 (on Pi 4) and the latest Raspberry Pi OS (on Zero W 2) on 2024-12-31.

1. Install libcomposites for the USB gadget mode. (From <https://github.com/milador/RaspberryPi-Joystick>)

        # Needed this on Raspberry Pi OS (?). Not needed on Ubuntu.
        echo "dtoverlay=dwc2" | sudo tee -a /boot/firmware/config.txt


        echo "dwc2" | sudo tee -a /etc/modules
        echo "libcomposite" | sudo tee -a /etc/modules

        reboot

1. Install binary on the Raspberry Pi.

        # Install commands
        apt install -y golang xxd git

        go install -v github.com/omakoto/raspberry-switch-control/nscontroller/cmd/...@latest

1. ~~Download source for the following script.~~

    This step is no longer needed. Now you can use the following sudo command to run the script.

1. Create the USB gadget by running the `switch-controller-gadget` script.

    The path to this script is not stable. Run `nsbackend usb-init-script-path` to get the path.

        sudo bash "$($HOME/go/bin/nsbackend usb-init-script-path)"

    Or, add the following entry to root's `crontabe` (i.e. `sudo crontab -e` and add it)
    Change `/home/pi/` as needed.


        @reboot bash -c ". $(/home/pi/go/bin/nsbackend usb-init-script-path)"


2. Connect the Raspberry Pi to the Nintendo Switch

  - If it's a Pi 4 or 5, use the USB-C port.
    
    My configuration: connect the Switch to a *powered* USB hub, then connect it to the Pi's C port. Make sure the Pi can draw enough power.

  - If it's a Zero, use the micro USB port.


## Control Nintendo Switch with Joystick on a PC (via Raspberry Pi)

1. Connect the Raspberry Pi to the Switch.
   - If using a Pi Zero, connect via the micro-USB port.
   - If using a Pi 4, connect to the USB C port. (aka the power port)
   - *Either way, to make sure the Pi keeps running even when not connected to the switch, use a powered USB hub.*

     (So, ideally, use a hub with a usb C output and connect it to the Pi, rather than using an A port.)

1. Connect a joystick to a host PC. (only the following ones are supported and tested)
    1. Nintendo Pro controller
    2. X-Box One controller
    3. PS4 controller

1. On the host PC, install the software:

        apt install -y golang 
        go install -v github.com/omakoto/raspberry-switch-control/nscontroller/cmd/...@latest

1. On the host PC, run it:

        nsfrontend -j /dev/input/js0 -o >(ssh pi@$PI_ADDRESS go/bin/nsbackend) 

1. Press `[enter]` on the console to finish.


## Run backend as daemon (Advanced use)

1. Auto start `nsbackend` as a daemon. Add this to `root`'s crontab.

        $ sudo crontab -l
        # ...
        @reboot bash -c ". $(/home/pi/go/bin/nsbackend usb-init-script-path); /home/pi/go/bin/nsbackend -x"


1. Then write to `/tmp/nsbackend.fifo` from `nsfrontend` instead:

        nsfrontend -j /dev/input/js0 -o >(ssh pi@$PI_ADDRESS 'echo "SSH Connected."; cat > /tmp/nsbackend.fifo') 


## TODOs

- Autofire on/off
- Macro


## References
1. Control switch from a smart phone
    1. https://mzyy94.com/blog/2020/03/20/nintendo-switch-pro-controller-usb-gadget/
    1. https://github.com/mzyy94/nscon
    1. https://gist.github.com/mzyy94/60ae253a45e2759451789a117c59acf9#file-add_procon_gadget-sh
1. https://www.kernel.org/doc/html/v4.13/driver-api/usb/gadget.html
1. https://github.com/milador/RaspberryPi-Joystick
1. https://www.rmedgar.com/blog/using-rpi-zero-as-keyboard-setup-and-device-definition
1. https://github.com/wchill/SwitchInputEmulator
1. https://sourceforge.net/projects/linuxconsole/
1. https://github.com/progmem/Switch-Fightstick
1. https://sourceforge.net/p/linuxconsole/code/ci/master/tree/utils/jstest.c
1. http://www.fourwalledcubicle.com/files/LUFA/Doc/120219/html/group___group___std_descriptors.html
