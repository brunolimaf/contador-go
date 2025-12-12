package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Estrutura para receber os dados de login do Javascript
type Credenciais struct {
	Usuario string `json:"usuario"`
	Senha   string `json:"senha"`
}

func main() {
	// 1. Configura o Go para servir os arquivos da pasta 'static' (HTML, CSS, JS)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 2. Cria uma rota específica apenas para o Login
	http.HandleFunc("/api/login", loginHandler)

	// === ALTERAÇÃO PARA O RENDER ===
	// Tenta pegar a porta que o Render definir
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Se não tiver (rodando local), usa a 8080
	}

	fmt.Println("Servidor rodando na porta:", port)

	// Atenção aqui: usamos ":" + port
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Função que processa o login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica se é um método POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodifica o JSON que vem do Javascript
	var creds Credenciais
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	// === LÓGICA DE LOGIN SIMPLES ===
	// Aqui definimos o usuário e senha corretos
	if creds.Usuario == "admin" && creds.Senha == "12345" {
		// Se correto, responde com sucesso
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
	} else {
		// Se incorreto, responde com erro
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": false})
	}
}
