package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Importa o driver do Postgres (o "_" é necessário)
)

// Estrutura para login
type Credenciais struct {
	Usuario string `json:"usuario"`
	Senha   string `json:"senha"`
}

// Variável global para conexão com o banco
var db *sql.DB

func main() {
	var err error

	// 1. Configuração da Conexão com o Banco
	// Pega a URL do Render OU usa a local do Docker
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Configuração local (igual ao docker-compose.yml)
		connStr = "postgres://user:password@localhost:5432/contador_db?sslmode=disable"
	}

	// Abre a conexão
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Testa a conexão
	if err = db.Ping(); err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}

	// Cria a tabela se não existir
	criarTabela()

	// 2. Configura Rotas
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/contador", contadorHandler) // Nova rota

	// 3. Inicia Servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Conectado ao DB! Servidor rodando na porta:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func criarTabela() {
	// Cria a tabela 'clicks' se não existir
	query := `CREATE TABLE IF NOT EXISTS clicks (
		id SERIAL PRIMARY KEY,
		quantidade INT NOT NULL DEFAULT 0
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	// Garante que existe a linha inicial com ID 1
	// (Tenta inserir, se já existir ID 1, não faz nada)
	db.Exec(`INSERT INTO clicks (id, quantidade) VALUES (1, 0) ON CONFLICT (id) DO NOTHING`)
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

// (Mantive seu loginHandler igual, apenas omiti para economizar espaço se já tiver)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método invalido", 405)
		return
	}
	var creds Credenciais
	json.NewDecoder(r.Body).Decode(&creds)
	if creds.Usuario == "admin" && creds.Senha == "12345" {
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
	} else {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]bool{"sucesso": false})
	}
}
