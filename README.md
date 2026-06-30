# aetherwave

**aetherwave** is a Software Defined Radio (SDR) receiver written in Go. This project is a learning laboratory designed to explore the fundamental concepts of Digital Signal Processing (**DSP**) and radio communications from the ground up.

## 🚀 Project Overview

The application interfaces with SDR hardware (specifically **RTL-SDR** based devices) to acquire raw I/Q signals, process them through a custom filtering and demodulation pipeline, and playback the resulting audio in real-time through the system's sound card.

## 🛠 Architecture & Organization

The project is modularly structured to separate hardware acquisition, mathematical processing, and output:

### 1. Core DSP (`internal/dsp`)
The mathematical heart of the project. It handles complex (I/Q) signals and implements:
*   **Signals & Samples:** Abstractions for complex samples (`Sample`) and signal buffers (`Signal`).
*   **Filtering:** **FIR** (Finite Impulse Response) filter implementation with coefficient calculation using Hamming windows.
*   **Fourier Transform (FFT):** Engine for spectral analysis, featuring both recursive and iterative in-place implementations.
*   **Resampling:** Logic to bridge the gap between high-frequency radio sample rates and standard audio rates.

### 2. Demodulation (`internal/demod`)
Algorithms to extract audio information from radio waves:
*   **AM (Amplitude Modulation):** Envelope extraction via magnitude calculation and DC-offset removal.
*   **SSB (Single Sideband):** Support for USB and LSB using heterodyning (frequency shifting via BFO) and real-part extraction.
*   **FM:** Interface prepared in the codebase.

### 3. Hardware & Drivers (`internal/hardware`)
Low-level communication with physical devices:
*   **RTL-SDR:** Integration with `librtlsdr` via CGO to configure center frequency, sample rate, and gain.
*   Uses a **buffer pool** strategy to minimize memory allocations and reduce garbage collector pressure.

### 4. Audio (`internal/audio`)
*   Manages acoustic output using the **PortAudio** library.
*   Implements asynchronous callbacks to ensure smooth playback without blocking the DSP pipeline.

### 5. Orchestration (`internal/radio` & `commands`)
*   **Receiver:** The "brain" that coordinates the pipeline. It manages concurrency using separate goroutines for hardware reading and signal processing.
*   **Commands:** Provides a Command Line Interface (CLI) to start listening on specific frequencies and modes.

## 📡 Processing Pipeline

The data flows as follows:
1.  **Hardware:** I/Q sampling from the RTL-SDR USB dongle.
2.  **DSP:** Low-pass filtering to isolate the target channel.
3.  **Demodulation:** Conversion of the radio signal into audio samples.
4.  **Resampling:** Adjusting the rate to match the sound card (e.g., 48kHz).
5.  **Output:** Real-time playback via PortAudio.

## 🔧 Requirements

*   **Go:** version 1.26+
*   **System Libraries:**
    *   `librtlsdr` (for SDR hardware access)
    *   `portaudio` (for audio output)
*   **Hardware:** An RTL-SDR compatible dongle (e.g., Nooelec, RTL-SDR Blog V3/V4).

## 🏁 Getting Started

*Note: Ensure the C libraries mentioned above are installed on your system.*

To start the radio on a specific frequency (e.g., 100.1 MHz in AM mode):

```aetherwave-cli listen -freq 100100000 -mode AM```
