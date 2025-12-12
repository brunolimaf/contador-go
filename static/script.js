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

    // Limpa mensagem de erro ao trocar de tela
    msgErro.innerText = ""; 

    if (modoAtual === 'login') {
        // Muda para modo Cadastro
        modoAtual = 'cadastro';
        titulo.innerText = "Criar Nova Conta";
        btn.innerText = "Cadastrar";
        btn.onclick = registrarUsuario; // Define a função do botão para Registrar
        texto.innerText = "Já tem conta?";
        link.innerText = "Fazer Login";
    } else {
        // Volta para modo Login
        modoAtual = 'login';
        titulo.innerText = "Acesso Restrito";
        btn.innerText = "Entrar";
        btn.onclick = fazerLogin; // Define a função do botão para Login
        texto.innerText = "Não tem conta?";
        link.innerText = "Cadastre-se";
    }
}

// === FUNÇÃO DE REGISTRO (NOVA) ===
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
            alternarModo(); // Volta automaticamente para a tela de login
        } else {
            // Tenta ler a mensagem de erro que o Go mandou (ex: "Usuário já existe")
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
            localStorage.setItem('usuario_logado', 'sim'); // Salva sessão no navegador
            mostrarApp();
            carregarContadorDoServidor(); // Busca o contador do banco
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
    
    // Se já estiver logado (pelo localStorage), pula o login
    if (logado === 'sim') {
        mostrarApp();
        carregarContadorDoServidor();
    }
};

// === FUNÇÕES DO CONTADOR (BANCO DE DADOS) ===

// Busca o valor atual (GET)
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

// Aumenta o valor (POST)
async function incrementar() {
    try {
        const res = await fetch('/api/contador', { method: 'POST' });
        const data = await res.json();
        
        // Atualiza a tela com o valor confirmado pelo banco
        contadorDisplay.innerText = data.valor;
    } catch (e) {
        console.error("Erro ao incrementar", e);
    }
}

// === FUNÇÕES DE TELA ===
function mostrarApp() {
    loginScreen.classList.add('hidden');
    appScreen.classList.remove('hidden');
}

function sair() {
    localStorage.removeItem('usuario_logado');
    location.reload(); 
}