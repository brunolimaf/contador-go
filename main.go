package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"        // Importa o driver do Postgres (o "_" é necessário)
	"golang.org/x/crypto/bcrypt" // <--- Biblioteca de Segurança
)

// Estrutura para login
type Credenciais struct {
	Usuario string `json:"usuario"`
	Senha   string `json:"senha"`
}

// Variável global para conexão com o banco
var db *sql.DB

func main() {
	// 1. Conexão com Banco (Igual a antes)
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
		log.Fatal("Erro ao conectar no banco:", err)
	}

	// Cria tabelas
	criarTabelas()

	// 2. Rotas
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/registrar", registrarHandler) // NOVA ROTA
	http.HandleFunc("/api/login", loginHandler)         // ROTA ATUALIZADA
	http.HandleFunc("/api/contador", contadorHandler)

	// 3. Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Servidor rodando na porta:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func criarTabelas() {
	// Tabela do Contador (Mantivemos)
	db.Exec(`CREATE TABLE IF NOT EXISTS clicks (
		id SERIAL PRIMARY KEY,
		quantidade INT NOT NULL DEFAULT 0
	);`)
	db.Exec(`INSERT INTO clicks (id, quantidade) VALUES (1, 0) ON CONFLICT (id) DO NOTHING`)

	// NOVA TABELA: Usuários
	// username deve ser UNIQUE para não ter dois iguais
	query := `CREATE TABLE IF NOT EXISTS usuarios (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Erro ao criar tabela usuarios:", err)
	}
}

// === NOVO: Registrar Usuário ===
func registrarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método invalido", 405)
		return
	}

	var creds Credenciais
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Erro no JSON", 400)
		return
	}

	// 1. Gerar o HASH da senha (Custo 10 é um bom balanço)
	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Senha), 10)
	if err != nil {
		http.Error(w, "Erro ao criptografar senha", 500)
		return
	}

	// 2. Salvar no Banco
	_, err = db.Exec("INSERT INTO usuarios (username, password_hash) VALUES ($1, $2)", creds.Usuario, string(hash))
	if err != nil {
		// Se der erro, provavelmente o usuário já existe
		http.Error(w, "Erro: Usuário provavelmente já existe", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"msg": "Usuário criado com sucesso!"})
}

// Handler para Ler (GET) e Atualizar (POST) o contador
func contadorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		// --- LER DO BANCO ---
		var qtd int
		// Pega o valor onde o ID é 1
		err := db.QueryRow("SELECT quantidade FROM clicks WHERE id = 1").Scan(&qtd)
		if err != nil {
			http.Error(w, "Erro ao ler banco", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]int{"valor": qtd})

	} else if r.Method == http.MethodPost {
		// --- ATUALIZAR NO BANCO ---
		// Atualiza somando +1 onde o ID é 1
		_, err := db.Exec("UPDATE clicks SET quantidade = quantidade + 1 WHERE id = 1")
		if err != nil {
			http.Error(w, "Erro ao atualizar banco", http.StatusInternalServerError)
			return
		}

		// Lê o novo valor para devolver ao front
		var novaQtd int
		db.QueryRow("SELECT quantidade FROM clicks WHERE id = 1").Scan(&novaQtd)
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
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Erro no JSON", 400)
		return
	}

	// 1. Buscar o HASH do usuário no banco
	var hashSalvo string
	err := db.QueryRow("SELECT password_hash FROM usuarios WHERE username=$1", creds.Usuario).Scan(&hashSalvo)

	if err == sql.ErrNoRows {
		// Usuário não encontrado
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": false})
		return
	}

	// 2. Comparar a senha enviada com o HASH salvo
	err = bcrypt.CompareHashAndPassword([]byte(hashSalvo), []byte(creds.Senha))

	if err == nil {
		// Senha bateu!
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
	} else {
		// Senha errada
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": false})
	}
}
