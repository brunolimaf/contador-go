// Elementos da tela
const loginScreen = document.getElementById('login-screen');
const appScreen = document.getElementById('app-screen');
const contadorDisplay = document.getElementById('contador');
const msgErro = document.getElementById('msg-erro');

// 1. Ao carregar a página, verifica se já existe sessão salva
window.onload = function() {
    const logado = localStorage.getItem('usuario_logado');
    
    if (logado === 'sim') {
        mostrarApp();
        carregarContador();
    }
};

// 2. Função para enviar dados ao Go e fazer Login
async function fazerLogin() {
    const usuarioInput = document.getElementById('username').value;
    const senhaInput = document.getElementById('password').value;

    // Envia os dados para o backend em Go
    const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ usuario: usuarioInput, senha: senhaInput })
    });

    if (response.ok) {
        // Se o Go disse que está OK:
        localStorage.setItem('usuario_logado', 'sim'); // Salva sessão
        mostrarApp();
        carregarContador();
        msgErro.innerText = "";
    } else {
        msgErro.innerText = "Usuário ou senha incorretos!";
    }
}

// 3. Função do Contador
function incrementar() {
    // Pega o valor atual
    let valorAtual = parseInt(localStorage.getItem('valor_contador') || 0);
    
    // Aumenta 1
    valorAtual++;
    
    // Salva no Local Storage (navegador)
    localStorage.setItem('valor_contador', valorAtual);
    
    // Atualiza a tela
    contadorDisplay.innerText = valorAtual;
}

// Auxiliares
function carregarContador() {
    // Busca do storage, se não tiver nada, assume 0
    const valorSalvo = localStorage.getItem('valor_contador') || 0;
    contadorDisplay.innerText = valorSalvo;
}

function mostrarApp() {
    loginScreen.classList.add('hidden');
    appScreen.classList.remove('hidden');
}

function sair() {
    // Limpa a sessão, mas opcionalmente mantemos o contador
    localStorage.removeItem('usuario_logado');
    location.reload(); // Recarrega a página
}