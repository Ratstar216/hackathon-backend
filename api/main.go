package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/ulid/v2"

	_ "github.com/go-sql-driver/mysql"
)

type UserResForHTTPGet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type RequestBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type TweetGet struct {
	Id      string `json:"id"`
	Userid  string `json:"userid"`
	Name    string `json:"name"`
	Times   string `json:"times"`
	Likes   int    `json:"likes"`
	Retweet int    `json:"retweet"`
	Content string `json:"content"`
}

type TweetPost struct {
	Userid  string `json:"userid"`
	Name    string `json:"name"`
	Times   string `json:"times"`
	Likes   int    `json:"likes"`
	Retweet int    `json:"retweet"`
	Content string `json:"content"`
}

type CL struct {
	Id     string `json:"id"`
	Userid string `json:"userid"`
}

// ① GoプログラムからMySQLへ接続
var db *sql.DB

func init() {
	// ①-1

	// err := godotenv.Load(".env")
	//if err != nil {
	//	log.Fatalf("fail: godotenv.Load, %v\n", err)
	//}

	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	mysqlHost := os.Getenv("MYSQL_HOST")
	// fmt.Println(mysqlUser, mysqlPwd, mysqlDatabase)

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	// fmt.Println(connStr)
	_db, err := sql.Open("mysql", connStr)

	// ①-2
	// _db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(localhost:3306)/%s", mysqlUser, mysqlPwd, mysqlDatabase))
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	// ①-3
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodGet:

		// ②-1
		// name := r.URL.Query().Get("name")
		// if name == "" {
		// 	log.Println("fail: name is empty")
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		// ②-2
		rows, err := db.Query("SELECT id, name, age FROM user")
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ②-3
		users := make([]UserResForHTTPGet, 0)
		for rows.Next() {
			var u UserResForHTTPGet
			if err := rows.Scan(&u.Id, &u.Name, &u.Age); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		// ②-4
		bytes, err := json.Marshal(users)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	case http.MethodPost:
		var requestBody RequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			log.Printf("fail: json.NewDecoder, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if requestBody.Name == "" || len(requestBody.Name) > 50 || requestBody.Age < 20 || requestBody.Age > 80 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		ins, err := tx.Prepare("INSERT INTO user VALUES (?, ?, ?)")
		if err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Prepare, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id := ulid.Make().String()
		_, err = ins.Exec(id, requestBody.Name, requestBody.Age)
		if err != nil {
			tx.Rollback()
			log.Printf("fail: ins.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Printf("{ \"id\" : \"%s\" }", id)
		response := map[string]interface{}{"id": id}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func tweethandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodGet:

		// ②-1
		// name := r.URL.Query().Get("name")
		// if name == "" {
		// 	log.Println("fail: name is empty")
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		// ②-2
		rows, err := db.Query("SELECT id, userid, name, times, likes, retweet, content FROM posts")
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ②-3
		tweets := make([]TweetGet, 0)
		for rows.Next() {
			var u TweetGet
			if err := rows.Scan(&u.Id, &u.Userid, &u.Name, &u.Times, &u.Likes, &u.Retweet, &u.Content); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			tweets = append(tweets, u)
		}

		// ②-4
		bytes, err := json.Marshal(tweets)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	case http.MethodPost:
		var requestBody TweetPost
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			log.Printf("fail: json.NewDecoder, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if requestBody.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		ins, err := tx.Prepare("INSERT INTO posts VALUES (?, ?, ?, ?, ?, ?, ?)")
		log.Printf("111111111111111")
		if err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Prepare, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id := ulid.Make().String()
		log.Printf("%v, %v, %v, %v, %v, %v, %v", id, requestBody.Userid, requestBody.Name, requestBody.Times, requestBody.Likes, requestBody.Retweet, requestBody.Content)
		_, err = ins.Exec(id, requestBody.Userid, requestBody.Name, requestBody.Times, requestBody.Likes, requestBody.Retweet, requestBody.Content)

		if err != nil {
			tx.Rollback()
			log.Printf("fail: ins.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Printf("{ \"id\" : \"%s\" }", id)
		response := map[string]interface{}{"id": id}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func likehandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodPost:
		var requestBody CL
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			log.Printf("fail: json.NewDecoder, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tx, err := db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		ins, err := tx.Prepare("UPDATE posts SET likes = likes + 1 WHERE id = (?)")
		log.Printf("111111111111111")
		if err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Prepare, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// id := ulid.Make().String()
		log.Printf("%v, %v", requestBody.Id, requestBody.Userid)
		_, err = ins.Exec(requestBody.Id)

		if err != nil {
			tx.Rollback()
			log.Printf("fail: ins.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Printf("{ \"id\" : \"%s\" }", id)
		//response := map[string]interface{}{"id": id}
		//w.Header().Set("Content-Type", "application/json")
		//json.NewEncoder(w).Encode(response)
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
	http.HandleFunc("/user", handler)

	http.HandleFunc("/tweet", tweethandler)

	http.HandleFunc("/posts", tweethandler)

	http.HandleFunc("/likes", likehandler)

	// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
	closeDBWithSysCall()

	// 8000番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}
