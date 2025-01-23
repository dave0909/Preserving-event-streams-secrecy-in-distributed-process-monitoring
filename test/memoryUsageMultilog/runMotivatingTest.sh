#!/bin/bash
#Motivating: 40740 + 2000 end events = 42740
# Function to run the process state agent
run_process_state_agent() {
    CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go run ../../procesStateAgent/processStateAgent.go localhost:6065 localhost:1234 false true
}

# Function to run the process vault
run_process_vault() {
    # Change to the directory where the main executable is located
    cd ../..
    python3 pv4.py ./data/PNML/motivating.pnml ./workflowLogic/workflowLogic.go ./data/regoConstraints/motivatingConstraints ./complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 data/input/extraction_manifest_motivating.json false true 42740 false 200
    # Change back to the original directory
    cd -
}

run_event_stream_generator(){
  cd ../../eventStreamGenerator
  python3 event_stream_from_log.py ../data/xes/motivatingnew.xes
  cd -
}

run_delay_hub() {
    CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go run delayHub.go localhost:8388
    }
# Trap SIGINT to terminate background processes
trap 'kill $(jobs -p); exit' SIGINT
kill -9 $(lsof -t -i:6066)



#Go to the directory where the delay hub is located
cd ../../delayHub
run_delay_hub &
#Change back to the original directory
cd -


sleep 2

kill -9 $(lsof -t -i:6066)

run_process_state_agent &
sleep 2

run_process_vault &
sleep 10

run_event_stream_generator &

# Wait for all background processes to finish
wait

