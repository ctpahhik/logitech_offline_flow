# "Offline" Logitech Flow Simulator

## Overview

Tool that emulates the core behavior of Logitech Flow: it automatically switches your keyboard and mouse control between multiple computers but does not require network connection between them.<br>
This allows seamless control across several machines as if you were using a single device, enhancing productivity and convenience in multi-device setups.<br>
Can be used instead of Logitech Flow when there is no direct network connection between devices due to various reasons (no network / VPN / corporate limitations etc).

## Features

- **Automatic Device Switching:** Instantly move your mouse pointer to the edge of one screen to control another computer.
- **Universal:** Should support any multihost Logi Bolt device in both Receiver and Bluetooth modes. Tested only for **MX Master 3S** Mouse and **K950** Keyboard.
- **Lightweight:** Minimal dependencies, fast startup.

## TODOs
- **Customizable Hotkeys:** Optionally switch devices using user-defined shortcuts.
- **Cross-Platform:** Work on Windows, Linux, and macOS (where supported).
- **Multimonitor setups:** You can choose either attach to the main screen or to whole desktop
- **Device lookup:** Enumerate devices for easier configuration
- **Non-Console app:** No Console screen. Simple tray icon.
  
## How It Works

Tool must be installed and manually configured (via command line parameters) on every host. When mouse pointer reaches left/right edge of the screen it sends 'change host' command to every device and moves pointer a little bit back to avoid flickering.
There is also small delay before next switch for the same reason.

## Usage

1. **Install** No installation required. Build / download the tool on all participating hosts and that's it.
2. **Start** the tool on each host, provide device indexes and target channels for the left and/or right edges
4. **Move** your mouse to the edge of your screen to switch control to another host!
5. **Hotkey** Press Ctrl+Shift+W to enable/disable automatic switching

## Configuration

**--devices=...** (mandatory) coma separated list of indexes of Logitech devices paired with the host. Usually index is the same as paired device position in any Logitech tool. Standard receiver support up to 6 devices (indexes 1-6) so if you don't know the index you can try all of them.<br>
**--left-channels=...** (optional) coma separated list of devices' channels paired with the receiver to be switched to when pointer touches left edge of the screen. Must be in the same order as devices.<br>
**--right-channels=...** (optional) coma separated list of devices' channels paired with the receiver to be switched to when pointer touches right edge of the screen. Must be in the same order as devices.<br>
At least one --left-channels or --right-channels must be provided.

## Example

hosts:
B - A - C

```sh
# Start on Host A (center)
offline_logi_flow_sim.exe --devices=1,2 --left-channels=0,0 --right-channels=2,2

# Start on Host B (left)
offline_logi_flow_sim.exe --devices=1,2 --right-channels=1,1

# Start on Host C (right)
offline_logi_flow_sim.exe --devices=1,2 --left-channels=1,1
```

- By moving your mouse cursor to the left edge of Host A’s screen control will switch to Host B, and vice versa, to the right edge of Host A’s screen, control will switch to Host C, and vice versa.

## Limitations

- Clipboard sharing, drag-and-drop, or file transfer features are **not** and won't be implemented in the simulator.
- Works only on Windows but Linux and/or macOs support may be added if needed.
- Works properly only for the main screen in multimonitor setup

## Contribution

Pull requests and suggestions are welcome!

## License

This project is licensed under the MIT License.

---
**Disclaimer:** This project is not affiliated with or endorsed by Logitech. It is an independent implementation inspired by Logitech Flow.

**This project is provided as is, without any warranty. Use it at your own risk.**
