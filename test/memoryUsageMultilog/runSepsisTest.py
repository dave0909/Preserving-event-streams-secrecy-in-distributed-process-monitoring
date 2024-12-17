import os
import threading

def main():
    def run_process_state_agent():
        os.system("CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib ego-go run ../../procesStateAgent/processStateAgent.go localhost:6065 localhost:1234 false")

    def run_proces_vault():
        os.system("python3 ../../pv3.py ../../data/BPMN/sepsis.bpmn ../../workflowLogic/workflowLogic.go ../../data/regoConstraints/sepsisConstraints ../../complianceCheckingLogic/complianceCheckingLogic.go localhost:6066 ../../data/input/extraction_manifest_sepsis.json true true 15200 false 150")

    psa_thread = threading.Thread(target=run_process_state_agent)
    #psa_thread.start()

    pv_thread = threading.Thread(target=run_proces_vault)
    pv_thread.start()

if __name__ == "__main__":
    main()