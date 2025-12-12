// === ELEMENTOS DA TELA ===
const loginScreen = document.getElementById('login-screen');
const appScreen = document.getElementById('app-screen');
const contadorDisplay = document.getElementById('contador');
const msgErro = document.getElementById('msg-erro');

// === LÓGICA DE TROCA DE TELA (LOGIN vs CADASTRO) ===
let modoAtual = 'login'; 

function alternarModo() {
    const titulo = document.getElementById('titulo-login');
    const btn = document.getElementById('btn-acao');
    const texto = document.getElementById('texto-alternar');
    const link = document.getElementById('link-alternar');

    msgErro.innerText = ""; 

    if (modoAtual === 'login') {
        modoAtual = 'cadastro';
        titulo.innerText = "Criar Nova Conta";
        btn.innerText = "Cadastrar";
        btn.onclick = registrarUsuario;
        texto.innerText = "Já tem conta?";
        link.innerText = "Fazer Login";
    } else {
        modoAtual = 'login';
        titulo.innerText = "Acesso Restrito";
        btn.innerText = "Entrar";
        btn.onclick = fazerLogin;
        texto.innerText = "Não tem conta?";
        link.innerText = "Cadastre-se";
    }
}

// === FUNÇÃO DE REGISTRO ===
async function registrarUsuario() {
    const usuarioInput = document.getElementById('username').value;
    const senhaInput = document.getElementById('password').value;

    if (!usuarioInput || !senhaInput) {
        msgErro.innerText = "Preencha todos os campos!";
        return;
    }

    try {
        const response = await fetch('/api/registrar', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ usuario: usuarioInput, senha: senhaInput })
        });

        if (response.ok) {
            alert("Conta criada com sucesso! Agora faça login.");
            alternarModo();
        } else {
            const data = await response.text(); 
            msgErro.innerText = "Erro: " + data;
        }
    } catch (e) {
        msgErro.innerText = "Erro ao conectar com servidor.";
        console.error(e);
    }
}

// === FUNÇÃO DE LOGIN ===
async function fazerLogin() {
    const usuarioInput = document.getElementById('username').value;
    const senhaInput = document.getElementById('password').value;

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ usuario: usuarioInput, senha: senhaInput })
        });

        if (response.ok) {
            localStorage.setItem('usuario_logado', 'sim');
            mostrarApp();
            carregarContadorDoServidor();
            msgErro.innerText = "";
        } else {
            msgErro.innerText = "Usuário ou senha incorretos!";
        }
    } catch (e) {
        msgErro.innerText = "Erro de conexão.";
        console.error(e);
    }
}

// === INICIALIZAÇÃO ===
window.onload = function() {
    const logado = localStorage.getItem('usuario_logado');
    if (logado === 'sim') {
        mostrarApp();
        carregarContadorDoServidor();
    }
};

// === FUNÇÕES DO CONTADOR ===
async function carregarContadorDoServidor() {
    try {
        const res = await fetch('/api/contador'); 
        const data = await res.json();
        console.log("Valor vindo do banco:", data.valor);
        contadorDisplay.innerText = data.valor;
    } catch (e) {
        console.error("Erro ao buscar contador", e);
    }
}

async function incrementar() {
    try {
        const res = await fetch('/api/contador', { method: 'POST' });
        const data = await res.json();
        contadorDisplay.innerText = data.valor;
    } catch (e) {
        console.error("Erro ao incrementar", e);
    }
}

function mostrarApp() {
    loginScreen.classList.add('hidden');
    appScreen.classList.remove('hidden');
}

function sair() {
    localStorage.removeItem('usuario_logado');
    location.reload(); 
}

// ==========================================
// === WEBSOCKET (TEMPO REAL) - O CÓDIGO FINAL ===
// ==========================================

console.log(">>> O SCRIPT CHEGOU NO FINAL E VAI TENTAR CONECTAR <<<");

// 1. Detecta protocolo
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${protocol}//${window.location.host}/ws`;

// 2. Conecta
console.log("Tentando conectar em:", wsUrl);
const socket = new WebSocket(wsUrl);

socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log("🔥 ATUALIZAÇÃO RECEBIDA:", data.valor);
    
    if (contadorDisplay) {
        contadorDisplay.innerText = data.valor;
    }
};

socket.onerror = (error) => {
    console.error("❌ ERRO NO WEBSOCKET:", error);
};

socket.onclose = () => {
    console.log("⚠️ WEBSOCKET DESCONECTADO");
};