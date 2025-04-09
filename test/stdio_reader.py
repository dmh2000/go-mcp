#!/usr/bin/env python3

import sys
import logging
import os

LOG_FILENAME = "stdio_reader_py.log"

def setup_logging():
    """Sets up logging to a file."""
    log_formatter = logging.Formatter(
        "%(asctime)s [%(levelname)s] %(filename)s:%(lineno)d - %(message)s"
    )
    root_logger = logging.getLogger()
    root_logger.setLevel(logging.INFO)

    # Ensure the directory for the log file exists
    log_dir = os.path.dirname(LOG_FILENAME)
    if log_dir and not os.path.exists(log_dir):
        try:
            os.makedirs(log_dir)
        except OSError as e:
            print(f"Error creating log directory {log_dir}: {e}", file=sys.stderr)
            # Fallback to stderr if directory creation fails
            handler = logging.StreamHandler(sys.stderr)
            handler.setFormatter(log_formatter)
            root_logger.addHandler(handler)
            logging.error("Could not create log directory. Logging to stderr.")
            return

    try:
        file_handler = logging.FileHandler(LOG_FILENAME, mode='a')
        file_handler.setFormatter(log_formatter)
        root_logger.addHandler(file_handler)
    except IOError as e:
        print(f"Error opening log file {LOG_FILENAME}: {e}", file=sys.stderr)
        # Fallback to stderr if file opening fails
        handler = logging.StreamHandler(sys.stderr)
        handler.setFormatter(log_formatter)
        root_logger.addHandler(handler)
        logging.error("Could not open log file. Logging to stderr.")


def main():
    """Reads stdin and logs the content."""
    setup_logging()
    logging.info("Starting up. Reading from stdin...")

    try:
        # Read all data from standard input
        # Use sys.stdin.buffer.read() to get bytes, then decode assuming UTF-8
        # Adjust encoding if necessary based on expected client output
        input_bytes = sys.stdin.buffer.read()
        input_data = input_bytes.decode('utf-8', errors='replace') # Decode bytes to string
        logging.info(f"Read {len(input_bytes)} bytes from stdin. Logging received data.")
        logging.info(f"Received Data:\n---\n{input_data}\n---")

    except Exception as e:
        logging.exception("An error occurred while reading or processing stdin.")
        sys.exit(1)

    logging.info("Finished processing stdin. Exiting.")

if __name__ == "__main__":
    main()
