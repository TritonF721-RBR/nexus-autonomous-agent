// Konfigurasi URL Golang kamu (sesuaikan jika portnya berbeda)
console.log("Server Terkoneksi");
const GOLANG_API_BASE = 'http://localhost:8080/task'; 

const chatBox = document.getElementById('chatBox');
const userInput = document.getElementById('userInput');

// Mengizinkan tombol 'Enter' untuk mengirim pesan
userInput.addEventListener('keypress', function (e) {
    if (e.key === 'Enter') {
        sendTask();
    }
});

// Fungsi Utama: Mengirim Tugas
async function sendTask() {
    const text = userInput.value.trim();
    if (!text) return;

    // 1. Tampilkan pesan user di layar
    appendMessage(text, 'user');
    userInput.value = '';

    // 2. Tampilkan pesan agen sedang "berpikir"
    const agentMessageElement = appendMessage('...', 'agent', true);

    try {
        // 3. Tembak POST ke Golang (seperti di Postman)
        const response = await fetch(GOLANG_API_BASE, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ instruction: text }) 
        });

        const data = await response.json();
        
        // 4. Dapatkan Task ID, lalu mulai Polling (ngecek status)
        if (data.task_id) {
            checkTaskStatus(data.task_id, agentMessageElement);
        } else {
            agentMessageElement.innerHTML = "Error: Tidak mendapatkan Task ID dari server.";
        }

    } catch (error) {
        agentMessageElement.innerHTML = "Error koneksi ke server Golang: " + error.message;
    }
}

// Fungsi Polling: Mengecek status ke Golang berulang kali
function checkTaskStatus(taskId, messageElement) {
    let pollingInterval = setInterval(async () => {
        try {
            // Tembak GET status
            const response = await fetch(`${GOLANG_API_BASE}/${taskId}/status`);
            const data = await response.json();

            if (data.status === 'COMPLETED') {
                // Hentikan interval (berhenti mengecek)
                clearInterval(pollingInterval);
                
                // Tampilkan hasil dari AI
                messageElement.classList.remove('processing-text');
                messageElement.innerHTML = data.result;
                scrollToBottom();
            } else {
                // Jika masih PROCESSING, biarkan efek berpikir
                messageElement.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i> Agen sedang memproses (' + data.status + ')...';
            }
        } catch (error) {
            clearInterval(pollingInterval);
            messageElement.innerHTML = "Error saat mengecek status: " + error.message;
        }
    }, 3000); // Cek setiap 3 detik
}

// Fungsi Bantuan: Menambahkan elemen obrolan ke HTML
function appendMessage(text, sender, isProcessing = false) {
    const messageDiv = document.createElement('div');
    messageDiv.classList.add('message', sender + '-message');

    const contentDiv = document.createElement('div');
    contentDiv.classList.add('message-content');
    
    if (isProcessing) {
        contentDiv.classList.add('processing-text');
        contentDiv.innerHTML = '<i class="fa-solid fa-circle-notch fa-spin"></i> Menganalisis tugas...';
    } else {
        contentDiv.textContent = text;
    }

    messageDiv.appendChild(contentDiv);
    chatBox.appendChild(messageDiv);
    scrollToBottom();

    return contentDiv; // Mengembalikan elemen agar teksnya bisa diupdate nanti
}

// Fungsi Bantuan: Selalu scroll ke bawah saat ada pesan baru
function scrollToBottom() {
    chatBox.scrollTop = chatBox.scrollHeight;
}