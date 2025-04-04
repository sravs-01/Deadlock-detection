# **RPC-Based Deadlock Detection Using Probe Messages and DDFS**  

This project implements a **distributed deadlock detection** algorithm using **Remote Procedure Calls (RPCs)** in **Go**. It uses **probe messages** and **Distributed Depth-First Search (DDFS)** to identify circular wait conditions in a distributed system.  

## **Features**  
✅ **Fully distributed**: Each process runs independently, communicating via **RPC over TCP**.  
✅ **Probe-based detection**: Sends probe messages to track dependencies and detect cycles.  
✅ **Session-based tracking**: Unique session IDs ensure correct message processing.  
✅ **DDFS Optimization**: Prevents redundant messages by avoiding revisited nodes.  
✅ **Real-time logging**: Logs message flow, timestamps, and deadlock detection events.  

---

## **How It Works**  

1. Each process has a list of **neighbor processes** that it depends on.  
2. A process that suspects a deadlock **sends a probe** to its neighbors via **RPC**.  
3. The probe tracks visited nodes, forming a **wait-for graph**.  
4. If a probe **returns to the origin**, a **deadlock is detected**.  
5. The system logs the detected cycle and stops the probe propagation.  

---

## **Installation & Setup**  

### **1. Clone the repository**  
```bash
git clone https://github.com/sravs-01/deadlock-detection.git
cd deadlock-detection
```

### **2. Run the RPC servers (Processes)**  
Each process runs as an independent RPC server. Open **multiple terminals** and start each process separately.

For example, to run **Process 1**:  
```bash
go run main.go
```

Do this for each process in the topology.

### **3. Start Deadlock Detection**  
Once all servers are running, **trigger deadlock detection** by sending an initial probe from Process 1. This is done automatically in the `main.go` file.

---

## **Example Output**  

```log
[P1] Listening on port 8001...
[P2] Listening on port 8002...
[P3] Listening on port 8003...
[P4] Listening on port 8004...

[P1] Sending probe to P2: Visited [1]
[P2] Sending probe to P3: Visited [1, 2]
[P3] Sending probe to P4: Visited [1, 2, 3]
[P4] Sending probe to P1: Visited [1, 2, 3, 4]

[P1] Deadlock detected! Cycle: [1, 2, 3, 4]
```

This means that a **circular wait** was detected involving processes 1 → 2 → 3 → 4 → 1.

---

## **Topology Configuration**  
The **process topology** is defined in `main.go` under:  

```go
processes := map[int][]int{
    1: {2},
    2: {3},
    3: {4},
    4: {1}, // Creates a cycle
}
```
Modify this to **test different scenarios**.

---

### **Deadlock Graph Representation**  
The following diagram illustrates the process dependencies and potential deadlock cycle:

![Deadlock Graph](images/deadlock_graph.png)

---

## 🖼️ Output Visualization

This section presents visual comparisons of the output from different versions of the deadlock detection system.

### 🔹 1. Original Version (Basic Detection)

This version demonstrates basic deadlock detection using static process dependencies.

![Original Version Output](images/original_output.png)

---

### 🔹 2. Enhanced Version Using Goroutines

This version uses Go **goroutines** for concurrency, enabling faster probe handling and non-blocking communication.

![Goroutines Output](images/goroutine_output.png)

---

### 🔹 3. Enhanced RPC Version (Server & Client)

This version uses **RPC** to simulate a truly distributed environment with **client-server** communication, accurate logging, and better message tracking.

#### 🔸 Server Output
Shows how each RPC server logs incoming and outgoing probe messages and detection events.

![RPC Server Output](images/rpc_server_output.png)

#### 🔸 Client Output
Displays how the initiating process sends the first probe and logs the traversal path.

![RPC Client Output](images/rpc_client_output.png)

---

## **Future Improvements**  
🔹 **Dynamic process addition/removal**  
🔹 **Fault tolerance with retry mechanisms**  
🔹 **Optimized message passing with gRPC**  

---

📄 Documentation
Offline versions are available in the docs/ folder (.pdf and .docx).

View the live version(s) here: [Google Docs Project Report](https://docs.google.com/document/d/1spu1dJ6mS8UE3ECQuBBQtrbaLZ5Qaf7oTRDXdGYRx0A/edit?usp=sharing)

---

## **License**  
This project is licensed under the **MIT License**.  

---

## 👥 Authors

- **N. Sai Kiran Varma** (CB.EN.U4CSE22424)  
- **S. Siva Pravallika** (CB.EN.U4CSE22440)  
- **Suman Panigrahi** (CB.EN.U4CSE22444) – [GitHub](https://github.com/suman1406)  
- **Sravani Orugranti** (CB.EN.U4CSE22457) – [GitHub](https://github.com/sravs-01)
