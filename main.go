package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket" // <--- NOVO: Biblioteca de WebSocket
	_ "github.com/lib/pq"          // Importa o driver do Postgres (o "_" é necessário)
	"golang.org/x/crypto/bcrypt"   // <--- Biblioteca de Segurança
)

// Estrutura para login
type Credenciais struct {
	Usuario string `json:"usuario"`
	Senha   string `json:"senha"`
}

// Variável global para conexão com o banco
var db *sql.DB

// === NOVO: Configuração do WebSocket ===
var clients = make(map[*websocket.Conn]bool) // Mapa de quem está conectado
var broadcast = make(chan int)               // Canal para enviar o número novo
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite conexão de qualquer lugar
	},
}

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@127.0.0.1:5432/contador_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Erro no DB:", err)
	}

	criarTabelas()

	// === NOVO: Inicia a "antena" que fica ouvindo mensagens para enviar ===
	go handleMessages()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/registrar", registrarHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/contador", contadorHandler)

	// === NOVO: Rota do WebSocket ===
	http.HandleFunc("/ws", wsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Servidor WebSocket rodando na porta:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// === NOVO: Transforma HTTP em WebSocket ===
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Atualiza a conexão para WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Deixa de fechar a conexão, mas garante que fecha se a função terminar
	defer ws.Close()

	// Registra o novo cliente
	clients[ws] = true

	// Loop infinito para manter a conexão viva
	for {
		var msg interface{}
		// Lê mensagens (mesmo que não usemos input do cliente por aqui, precisamos ler)
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

// === NOVO: Envia mensagens para todos os conectados ===
func handleMessages() {
	for {
		// Espera chegar um número novo
		novoValor := <-broadcast

		// --- LOG DE DEBUG ---
		//fmt.Println("Backend: Recebi novo valor para espalhar:", novoValor)
		//fmt.Println("Backend: Tenho", len(clients), "clientes conectados.")
		// --------------------

		for client := range clients {
			err := client.WriteJSON(map[string]int{"valor": novoValor})
			if err != nil {
				fmt.Println("Erro ao enviar para cliente, desconectando...")
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func criarTabelas() {
	db.Exec(`CREATE TABLE IF NOT EXISTS clicks (id SERIAL PRIMARY KEY, quantidade INT NOT NULL DEFAULT 0);`)
	db.Exec(`INSERT INTO clicks (id, quantidade) VALUES (1, 0) ON CONFLICT (id) DO NOTHING`)
	db.Exec(`CREATE TABLE IF NOT EXISTS usuarios (id SERIAL PRIMARY KEY, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL);`)
}

// === NOVO: Registrar Usuário ===
func registrarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método invalido", 405)
		return
	}
	var creds Credenciais
	json.NewDecoder(r.Body).Decode(&creds)
	hash, _ := bcrypt.GenerateFromPassword([]byte(creds.Senha), 10)
	_, err := db.Exec("INSERT INTO usuarios (username, password_hash) VALUES ($1, $2)", creds.Usuario, string(hash))
	if err != nil {
		http.Error(w, "Erro registro", 400)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"msg": "Criado"})
}

// Handler para Ler (GET) e Atualizar (POST) o contador
func contadorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		var qtd int
		db.QueryRow("SELECT quantidade FROM clicks WHERE id = 1").Scan(&qtd)
		json.NewEncoder(w).Encode(map[string]int{"valor": qtd})

	} else if r.Method == http.MethodPost {
		// 1. Atualiza no Banco
		_, err := db.Exec("UPDATE clicks SET quantidade = quantidade + 1 WHERE id = 1")
		if err != nil {
			http.Error(w, "Erro DB", 500)
			return
		}

		// 2. Pega o novo valor
		var novaQtd int
		db.QueryRow("SELECT quantidade FROM clicks WHERE id = 1").Scan(&novaQtd)

		// 3. === NOVO: Avisa o canal de Broadcast ===
		// Isso vai acionar a função handleMessages lá em cima
		broadcast <- novaQtd

		// 4. Responde para quem clicou (o HTTP padrão)
		json.NewEncoder(w).Encode(map[string]int{"valor": novaQtd})
	}
}

// === ATUALIZADO: Login Seguro ===
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método invalido", 405)
		return
	}
	var creds Credenciais
	json.NewDecoder(r.Body).Decode(&creds)
	var hashSalvo string
	err := db.QueryRow("SELECT password_hash FROM usuarios WHERE username=$1", creds.Usuario).Scan(&hashSalvo)
	if err == sql.ErrNoRows {
		w.WriteHeader(401)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hashSalvo), []byte(creds.Senha)) == nil {
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
	} else {
		w.WriteHeader(401)
	}
}
