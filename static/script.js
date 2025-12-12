// === ELEMENTOS DA TELA ===
const loginScreen = document.getElementById('login-screen');
const appScreen = document.getElementById('app-screen');
const contadorDisplay = document.getElementById('contador');
const msgErro = document.getElementById('msg-erro');

// === LÓGICA DE TROCA DE TELA ===
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

// === REGISTRO ===
async function registrarUsuario() {
    const usuarioInput = document.getElementById('username').value;
    const senhaInput = document.getElementById('password').value;

    if (!usuarioInput || !senhaInput) { msgErro.innerText = "Preencha todos os campos!"; return; }

    try {
        const response = await fetch('/api/registrar', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ usuario: usuarioInput, senha: senhaInput })
        });
        if (response.ok) {
            alert("Conta criada! Faça login.");
            alternarModo();
        } else {
            const data = await response.text(); 
            msgErro.innerText = "Erro: " + data;
        }
    } catch (e) { console.error(e); }
}

// === LOGIN ===
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
            msgErro.innerText = "Dados incorretos!";
        }
    } catch (e) { console.error(e); }
}

// === INICIALIZAÇÃO ===
window.onload = function() {
    const logado = localStorage.getItem('usuario_logado');
    if (logado === 'sim') {
        mostrarApp();
        carregarContadorDoServidor();
    }
    // Inicia o WebSocket independente do login para garantir conexão
    conectarWebSocket();
};

// === API CONTADOR ===
async function carregarContadorDoServidor() {
    try {
        const res = await fetch('/api/contador'); 
        const data = await res.json();
        contadorDisplay.innerText = data.valor;
    } catch (e) { console.error(e); }
}

async function incrementar() {
    try {
        const res = await fetch('/api/contador', { method: 'POST' });
        const data = await res.json();
        contadorDisplay.innerText = data.valor;
    } catch (e) { console.error(e); }
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
// === WEBSOCKET ROBUSTO (AUTO-RECONNECT) ===
// ==========================================

let socket; // Variável global para o socket

function conectarWebSocket() {
    // Evita criar duplicatas se já estiver conectado
    if (socket && socket.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    console.log("Tentando conectar WebSocket...");
    socket = new WebSocket(wsUrl);

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (contadorDisplay) {
                contadorDisplay.innerText = data.valor;
            }
        } catch (e) {}
    };

    socket.onopen = () => {
        console.log("WebSocket Conectado!");
    };

    // A MÁGICA ESTÁ AQUI: Se cair, tenta de novo em 2 segundos
    socket.onclose = () => {
        console.log("WebSocket caiu. Reconectando em 2 segundos...");
        setTimeout(conectarWebSocket, 2000);
    };

    socket.onerror = (err) => {
        console.error("Erro WS, vai fechar e tentar reconectar.");
        socket.close(); // Força o fechamento para disparar o onclose
    };
}