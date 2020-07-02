# raspberry-switch-control
Emulate Nintendo Switch USB Controller with Raspberry Pi

Just after started the project, I found he following project, which almost solves my project. My project is going to be based on it.

1. Control switch from a smart phone
    1. https://mzyy94.com/blog/2020/03/20/nintendo-switch-pro-controller-usb-gadget/
    1. https://github.com/mzyy94/nscon
    1. https://gist.github.com/mzyy94/60ae253a45e2759451789a117c59acf9#file-add_procon_gadget-sh


## Hardware requirements
* Raspberry Pi Zero W (Supposedly, it works with a regular pi 4, via the USB-C port.)

## Set up Raspberry Pi Zero
From <https://github.com/milador/RaspberryPi-Joystick>:

1. Install libcomposite for the USB gadget mode:

        echo "dtoverlay=dwc2" | sudo tee -a /boot/config.txt
        echo "dwc2" | sudo tee -a /etc/modules
        echo "libcomposite" | sudo tee -a /etc/modules
1. 




## References
1. <https://www.kernel.org/doc/html/v4.13/driver-api/usb/gadget.html>
1. <https://github.com/milador/RaspberryPi-Joystick>
1. <https://www.rmedgar.com/blog/using-rpi-zero-as-keyboard-setup-and-device-definition>
1. <https://github.com/wchill/SwitchInputEmulator>
1. <https://sourceforge.net/projects/linuxconsole/>
1. <https://github.com/progmem/Switch-Fightstick>
1. <https://sourceforge.net/p/linuxconsole/code/ci/master/tree/utils/jstest.c>
1. <http://www.fourwalledcubicle.com/files/LUFA/Doc/120219/html/group___group___std_descriptors.html>
