# raspberry-switch-control

Emulate Nintendo Switch USB Controller with Raspberry Pi

Credit: This project heavily relies on https://github.com/mzyy94/nscon.

## Hardware requirements
* Raspberry Pi Zero W (Supposedly, it works with a regular pi 4, via the USB-C port.)

## Set up Raspberry Pi Zero

1. Install libcomposites for the USB gadget mode. (From <https://github.com/milador/RaspberryPi-Joystick>)

        echo "dtoverlay=dwc2" | sudo tee -a /boot/config.txt
        echo "dwc2" | sudo tee -a /etc/modules
        echo "libcomposite" | sudo tee -a /etc/modules

1. Install on the raspberry pi

        go get -v -u -t github.com/omakoto/raspberry-switch-control/nscontroller/...

1. Create the USB gadget. Do it once after every reboot.

        sudo go/src/github.com/omakoto/raspberry-switch-control/scripts/switch-controller-gadget

## Control Nintendo Switch with Joystick on a PC (via Raspberry Pi)

1. Plug in the Raspberry Pi to the Switch. If using a Pi Zero, just connect via the micro-USB cable.

1. Connect a joystick (only the following ones are supported and tested) to a host PC.
    1. Nintendo Pro controller
    1. X-Box One controller
    1. PS4 controller

1. On the host PC, install the software:

        go get -v -u -t github.com/omakoto/raspberry-switch-control/nscontroller/...

1. On the host PC, run it:

        nsfrontend -j /dev/input/js0 -o >(ssh pi@$PI_ADDRESS go/bin/nsbackend) 

1. Press `[enter]` on console to finish.

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
