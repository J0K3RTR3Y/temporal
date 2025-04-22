#!/usr/bin/env python3
"""
Simple Quantum MIDI Processor
-----------------------------

A minimal implementation of the quantum-inspired MIDI processing algorithm.
"""

import os
import time
import random
import numpy as np
from pathlib import Path

def setup_logging():
    """Initialize logging and create output directory."""
    log_file = "quantum_midi.log"
    with open(log_file, "w") as f:
        f.write("QUANTUM MIDI PROCESSOR LOG\n")
        f.write("=========================\n\n")
        f.write(f"Started at: {time.asctime()}\n\n")
    
    output_dir = Path("quantum_midi_output")
    if not output_dir.exists():
        os.makedirs(output_dir)
        log(f"Created output directory: {output_dir}")
    
    return log_file, output_dir

def log(message, file="quantum_midi.log"):
    """Write message to log file and print to console."""
    print(message)
    with open(file, "a") as f:
        f.write(message + "\n")

def generate_sample_data():
    """Generate sample MIDI data based on C major scale."""
    c_major = [60, 62, 64, 65, 67, 69, 71, 72]
    notes = [random.choice(c_major) for _ in range(20)]
    return notes

def quantum_transform(midi_data):
    """Apply quantum-inspired transformation to MIDI data."""
    # Create quantum state
    q_state = np.zeros(128, dtype=np.complex64)
    for note in midi_data:
        q_state[note] = 1.0
    
    # Normalize
    norm = np.linalg.norm(q_state)
    if norm > 0:
        q_state = q_state / norm
    
    # Apply Hadamard-like transformation
    hadamard = 1/np.sqrt(2) * np.array([[1, 1], [1, -1]], dtype=np.complex64)
    
    for i in range(0, 120, 2):
        if q_state[i] != 0 or (i+1 < 128 and q_state[i+1] != 0):
            # Get 2-note subspace
            subspace = np.array([q_state[i], q_state[i+1 if i+1 < 128 else 0]])
            # Apply transformation
            transformed = np.dot(hadamard, subspace)
            # Update state
            q_state[i] = transformed[0]
            if i+1 < 128:
                q_state[i+1] = transformed[1]
    
    # Measure the quantum state
    probabilities = np.abs(q_state) ** 2
    threshold = 0.1
    new_notes = np.where(probabilities > threshold)[0]
    
    return new_notes

def save_results(output_dir, original_notes, transformed_notes):
    """Save processing results to file."""
    result_file = output_dir / "quantum_result.txt"
    try:
        with open(result_file, "w") as f:
            f.write("Quantum MIDI Processing Result\n")
            f.write("============================\n\n")
            f.write(f"Original notes: {original_notes}\n\n")
            f.write(f"Transformed notes: {transformed_notes}\n")
        log(f"Results saved to {result_file}")
    except Exception as e:
        log(f"Error saving results: {e}")

def main():
    """Main processing function."""
    log_file, output_dir = setup_logging()
    
    # Generate and process MIDI data
    log("\nProcessing MIDI data with quantum-inspired algorithms")
    notes = generate_sample_data()
    log(f"Generated {len(notes)} MIDI notes")
    log(f"Notes: {notes}\n")
    
    # Apply quantum transformation
    log("Applying quantum transformations...")
    new_notes = quantum_transform(np.array(notes))
    log(f"Transformation complete. Generated {len(new_notes)} new notes")
    log(f"New notes: {new_notes}\n")
    
    # Save results
    save_results(output_dir, notes, new_notes)
    
    log("Quantum MIDI processing completed successfully!")
    log(f"Finished at: {time.asctime()}")
    log("\nCHECK THE OUTPUT DIRECTORY FOR RESULTS")
    
    print(f"\nQuantum MIDI processor finished. See {log_file} and the quantum_midi_output directory for results.")

if __name__ == "__main__":
    main()