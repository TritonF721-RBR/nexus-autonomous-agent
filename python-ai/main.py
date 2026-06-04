import os
import requests
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

# Menggunakan metode klasik yang elegan dari LangChain 0.1.20
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain.agents import initialize_agent, AgentType, Tool
from dotenv import load_dotenv

load_dotenv(dotenv_path="../.env")

app = FastAPI(title="Enterprise Autonomous AI Agent (Python Brain)")

# ==========================================
# 1. MEMUAT INFRASTRUKTUR DASAR
# ==========================================
print("⏳ Sedang memuat model AI HuggingFace...")
embedding_model = SentenceTransformer('all-MiniLM-L6-v2')
print("✅ Model Embedding Siap!")

print("⏳ Sedang memuat LLM Gemini...")
llm = ChatGoogleGenerativeAI(
    model="gemini-2.5-flash",  
    temperature=0, 
)

# ==========================================
# 2. DEFINISI TOOLS / ALAT BANTU 
# ==========================================
def database_search_tool(query: str) -> str:
    """Berguna untuk mencari data spesifik atau dokumen internal perusahaan dari Database."""
    try:
        golang_url = "http://localhost:8080/api/internal/search"
        payload = {"query": query}
        
        response = requests.post(golang_url, json=payload, timeout=5)
        
        if response.status_code == 200:
            data = response.json()
            return f"Ditemukan dokumen: {data.get('content', 'Tidak ada isi')}"
        else:
            return "Maaf, tidak ditemukan data yang relevan di database internal."
    except Exception as e:
        return f"Error saat menghubungi database: {str(e)}"

def calculator_tool(expression: str) -> str:
    """Berguna untuk menghitung operasi matematika yang rumit."""
    try:
        return str(eval(expression))
    except Exception as e:
        return f"Error perhitungan: {str(e)}"

# Daftarkan tools
tools = [
    Tool(
        name="DatabaseSearch",
        func=database_search_tool,
        description="Gunakan ini HANYA JIKA kamu perlu mencari data rahasia/internal perusahaan."
    ),
    Tool(
        name="Calculator",
        func=calculator_tool,
        description="Gunakan ini untuk menghitung angka atau rumus matematika."
    )
]

# ==========================================
# 3. INISIALISASI AGEN
# ==========================================
agent = initialize_agent(
    tools, 
    llm, 
    agent=AgentType.ZERO_SHOT_REACT_DESCRIPTION, 
    verbose=True,
    handle_parsing_errors=True,
    max_iterations=3
)
print("✅ AI Agent Otonom Siap Menerima Tugas!")


# ==========================================
# REQUEST MODELS & ENDPOINTS
# ==========================================
class TextRequest(BaseModel):
    text: str

class AgentTaskRequest(BaseModel):
    task_id: str
    instruction: str

@app.post("/api/embed")
async def generate_embedding(request: TextRequest):
    vector_array = embedding_model.encode(request.text).tolist()
    return {
        "dimension": len(vector_array),
        "embedding": vector_array
    }

@app.post("/api/agent/execute")
async def execute_agent_task(request: AgentTaskRequest):
    print(f"🚀 Memulai tugas {request.task_id}: {request.instruction}")
    
    try:
        result = agent.invoke({"input": request.instruction})
        final_answer = result.get("output")
    except Exception as e:
        final_answer = f"Agen mengalami kegagalan sistem: {str(e)}"

    return {
        "task_id": request.task_id,
        "status": "COMPLETED",
        "result": final_answer
    }