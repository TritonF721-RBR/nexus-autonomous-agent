import os
import google.generativeai as genai
from dotenv import load_dotenv

# Memuat Kunci dari file .env
load_dotenv()
api_key = os.getenv("GOOGLE_API_KEY")

if not api_key:
    print("❌ API Key tidak ditemukan!")
else:
    genai.configure(api_key=api_key)
    print("🔍 Memeriksa model yang tersedia untuk API Key kamu...\n")
    
    try:
        for m in genai.list_models():
            if 'generateContent' in m.supported_generation_methods:
                # Menghapus kata 'models/' di depannya agar sesuai dengan format LangChain
                print(f"✅ Nama Model yang sah: {m.name.replace('models/', '')}")
    except Exception as e:
        print(f"❌ Terjadi error saat menghubungi Google: {e}")