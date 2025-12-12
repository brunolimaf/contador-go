// Elementos da tela
const loginScreen = document.getElementById('login-screen');
const appScreen = document.getElementById('app-screen');
const contadorDisplay = document.getElementById('contador');
const msgErro = document.getElementById('msg-erro');

window.onload = function() {
  const logado = localStorage.getItem('usuario_logado');
  if (logado === 'sim') {
      mostrarApp();
      carregarContadorDoServidor(); // <--- TEM QUE SER ESSA NOVA
  }
};

async function fazerLogin() {
  const usuarioInput = document.getElementById('username').value;
  const senhaInput = document.getElementById('password').value;

  const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ usuario: usuarioInput, senha: senhaInput })
  });

  if (response.ok) {
      localStorage.setItem('usuario_logado', 'sim');
      mostrarApp();
      
      // === A CORREÇÃO É AQUI ===
      // Antes estava: carregarContador();
      // Agora deve ser:
      carregarContadorDoServidor(); 
      // =========================
      
      msgErro.innerText = "";
  } else {
      msgErro.innerText = "Usuário ou senha incorretos!";
  }
}

// NOVA FUNÇÃO: Manda incrementar no Banco de Dados
async function incrementar() {
  try {
      const res = await fetch('/api/contador', { method: 'POST' }); // Faz um POST
      const data = await res.json();
      
      // Atualiza a tela com o valor que voltou do banco
      contadorDisplay.innerText = data.valor;
  } catch (e) {
      console.error("Erro ao incrementar", e);
  }
}

// NOVA FUNÇÃO: Busca o valor atual do Banco de Dados
async function carregarContadorDoServidor() {
  try {
      const res = await fetch('/api/contador'); // Faz o GET
      const data = await res.json();
      console.log("Valor vindo do banco:", data.valor); // <--- Adicionei esse log para debug
      contadorDisplay.innerText = data.valor;
  } catch (e) {
      console.error("Erro ao buscar contador", e);
  }
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