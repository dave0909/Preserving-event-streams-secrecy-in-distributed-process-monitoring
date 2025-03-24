
# ProMTEE: Preserving data secrecy for online process monitoring

This repository contains the published prototype of ProMTEE, a framework to preserve data secrecy for online process monitoring. ProMTEE leverages trusted applications running in Intel SGX TEEs to execute control flow and compliance checking monitoring tasks, while hiding sensitive information of the input events.

## Repository Structure

The repository is organized into the following directories:

- `data`: Contains BPMN, PNML, and XES files, as well as input and output data.
- `eventStreamGenerator`: Contains Python scripts for generating and processing event logs.
- `processVault`: Contains Go files for compliance checking logic, event dispatching, and process state management.
- `utils`: Contains utility Go files for attestation, delay arguments, event submission, and Petri net operations.
- `testConfigurations`: Contains scripts for running tests in different modes.

## Setup and Running the Project

### Dependencies

To set up and run the project, you need to have the following dependencies installed:

- Go (version 1.16 or later)
- Python (version 3.6 or later)
- pm4py (Python library for process mining)
- ego (Edgeless Systems' confidential computing framework)
- INTEL SGX enabled CPU (required to run the process vault in non-simulation mode)

### Environment Setup

1. Clone the repository:

   ```sh
   git clone https://github.com/dave0909/Preserving-event-streams-secrecy-in-distributed-process-monitoring.git
   cd Preserving-event-streams-secrecy-in-distributed-process-monitoring
   ```

2. Install Go dependencies:

   ```sh
   go mod tidy
   ```

3. Install Python dependencies:

   ```sh
   pip install pm4py
   ```

4. Set up ego (refer to the official documentation for detailed instructions):


### Running the Project

1. Start the process state agent:

   ```sh
   cd procesStateAgent
   ego-go run processStateAgent.go localhost:6065 localhost:1234 false true
   ```

   Parameters for running the process state agent:
   - `psaServer`: The address to bind the RPC server.
   - `esgAddress`: The address of the event stream generator.
   - `skippAttestation`: Boolean indicating whether to skip attestation.
   - `testMode`: Boolean indicating whether to run in test mode.

2. Start the process vault:

   ```sh
   cd processVault
   python3 processVaultCompiler.py ./data/PNML/motivatingreduced.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/motivatingConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_motivating.json false true 40000 false 200
   ```

   Parameters for running the process vault:
   - `bpmn_file_path`: The path to the BPMN or PNML file.
   - `output_go_file_path`: The path to the output Go file for the workflow logic.
   - `constraint_folder_path`: The path to the folder containing the Rego constraint files.
   - `output_go_file_path_compliance`: The path to the output Go file for the compliance checking logic.
   - `event_dispatcher_address`: The address to bind the RPC server for the event dispatcher.
   - `extraction_manifest_file_path`: The path to the extraction manifest file.
   - `isInSimulation`: Boolean indicating whether to run in simulation mode.
   - `isInTesting`: Boolean indicating whether to run in test mode.
   - `nEvents`: The number of events to process.
   - `withExternalQueue`: Boolean indicating whether to enable external query.
   - `slidingWindowSize`: The size of the sliding window for event processing.

3. Start the event stream generator:

   ```sh
   cd eventStreamGenerator
   python3 event_stream_from_log.py ../data/xes/motivatingnew.xes
   ```

   Parameters for running the event stream generator:
   - `log_path`: The path to the XES event log file.

## Running Tests

### Test Configurations

The repository includes scripts for running tests in different modes, located in the `testConfigurations` directory. The available test modes are:

- `simulationMode`: Scripts for running tests in simulation mode.
- `teeMode`: Scripts for running tests in Trusted Execution Environment (TEE) mode.

### Running a Test

To run a test, navigate to the appropriate test mode directory and execute the desired test script. For example, to run the BPIC2012 test in TEE mode:

```sh
cd testConfigurations/teeMode
./runBPIC2012.sh
```

## Directory Contents

### `data`

Contains BPMN, PNML, and XES files, as well as input and output data.

### `eventStreamGenerator`

Contains Python scripts for generating and processing event logs.

### `processVault`

Contains Go files for compliance checking logic, event dispatching, and process state management.

### `utils`

Contains utility Go files for attestation, delay arguments, event submission, and Petri net operations.

### `testConfigurations`

Contains scripts for running tests in different modes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
