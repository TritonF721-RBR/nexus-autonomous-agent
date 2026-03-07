from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

# Inisialisasi API Framework
app = FastAPI(title="Enterprise Local AI Embedding Service")

# Memuat model AI ke dalam RAM (Akan men-download ~80MB saat pertama kali dijalankan)
print("⏳ Sedang memuat model AI HuggingFace...")
model = SentenceTransformer('all-MiniLM-L6-v2')
print("✅ Model AI Siap Bertempur!")

# Format JSON yang diharapkan dari Golang
class TextRequest(BaseModel):
    text: str

# Endpoint untuk menerima teks dari Golang
@app.post("/api/embed")
async def generate_embedding(request: TextRequest):
    # Mengubah teks manusia menjadi deretan 384 angka matematis (Vector)
    vector_array = model.encode(request.text).tolist()
    
    return {
        "dimension": len(vector_array),
        "embedding": vector_array
    }