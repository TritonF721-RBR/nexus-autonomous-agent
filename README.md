# 🚀 Nexus: Autonomous AI Agent with Decoupled Microservices

![CI/CD Status](https://github.com/TritonF721-RBR/nexus-autonomous-agent/actions/workflows/ci.yml/badge.svg)

## 📖 Project Overview
Nexus is an enterprise-grade Autonomous AI Agent prototype engineered to solve the classic latency and UI-blocking issues commonly found in standard RAG (Retrieval-Augmented Generation) applications. It empowers the AI not just to answer questions, but to autonomously decide when to use specific tools (like a mathematical calculator or vector database retrieval) to accomplish user requests.

## 💡 The Problem: Monolithic Bottlenecks
In typical monolithic AI applications, when the backend is processing heavy computations or scanning massive vector databases, the HTTP connection to the client remains open and blocked. If the processing time exceeds standard limits, the browser triggers a connection timeout, the UI freezes, and the user experience is entirely compromised. 

## 🧠 The Engineering Solution: Decoupled Architecture
To eliminate timeouts and ensure high availability, Nexus implements a decoupled microservices architecture combined with an Asynchronous Task Polling mechanism. The workload is strictly separated based on language strengths:

1. **The Muscle (Golang API Gateway):** Acts as a highly concurrent entry point. It instantly receives user prompts, generates a unique `Task ID`, and delegates the heavy lifting to the AI worker. It acts as a non-blocking status checker for the frontend.
2. **The Brain (Python AI Core):** A dedicated, isolated AI worker built with FastAPI and LangChain. It focuses purely on cognitive tasks: reasoning, autonomous tool-calling, and generating LLM responses.
3. **The Vault (PostgreSQL + pgvector):** Containerized via Docker, serving as a robust and persistent vector database for the RAG semantic search operations.
4. **Asynchronous UI (Frontend):** Instead of waiting for a single delayed HTTP response, the Vanilla JS frontend polls the Golang gateway using the `Task ID` every 3 seconds. This provides a seamless, real-time loading animation without ever dropping the connection.

## 🛠️ Tech Stack
* **API Gateway (The Muscle):** Golang (v1.21)
* **AI Core (The Brain):** Python (v3.10), FastAPI, LangChain
* **Vector Database (The Vault):** PostgreSQL with `pgvector` (Dockerized)
* **Frontend:** Vanilla HTML, CSS, JavaScript (Asynchronous Polling)
* **CI/CD Pipeline:** GitHub Actions

## 🚀 Local Installation & Execution Guide

### Prerequisites
Before diving in, ensure your machine is equipped with the following:
* [Golang](https://go.dev/doc/install) (v1.21 or higher)
* [Python](https://www.python.org/downloads/) (v3.10 or higher)
* [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Required to run the PostgreSQL container)
* A valid **Gemini API Key** from [Google AI Studio](https://aistudio.google.com/)

### Step 1: Clone the Repository & Configure Secrets
```bash
git clone [https://github.com/TritonF721-RBR/nexus-autonomous-agent.git](https://github.com/TritonF721-RBR/nexus-autonomous-agent.git)
cd nexus-autonomous-agent
```
*Note: Create a `.env` file in both the `golang-api` and `python-ai` directories. Populate them with your Database credentials and Gemini API Key to prevent credential leaks.*

### Step 2: Ignite the Vector Database (Docker)
Ensure Docker Desktop is running in the background, then execute the following command in the root directory:
```bash
docker-compose up -d
```

### Step 3: Boot Up the AI Core (Python)
Open a new terminal, navigate to the Python directory, establish the environment, and start the AI engine:
```bash
cd python-ai
python -m venv venv

# For Windows:
venv\Scripts\activate
# For Mac/Linux:
source venv/bin/activate

pip install -r requirements.txt
uvicorn main:app --port 8001 --reload
```

### Step 4: Boot Up the API Gateway (Golang)
Open another terminal tab, navigate to the Golang directory, and start the gateway:
```bash
cd golang-api
go mod tidy
go run main.go
```
*The Golang server is now actively listening and routing requests on `http://localhost:8080`.*

### Step 5: Launch the Asynchronous UI
Use a local web server (such as the **Live Server** extension in VS Code) to serve the `index.html` file located in the `front-end` folder. Type a complex prompt (e.g., a mathematical calculation or document search) and watch the AI autonomously select the correct tools without blocking your UI!

## 🛡️ Future Improvements & Production Readiness
To maintain focus on the core decoupled AI Agent architecture, certain production-level features (such as Authentication and Rate Limiting) have been intentionally omitted from this prototype. For a full production deployment, the following enhancements are planned:
* **JWT (JSON Web Token) Middleware:** Implementation at the Golang API Gateway level to secure endpoints.
* **IP-based Rate Limiting:** To prevent spam, DDoS attacks, and API abuse.