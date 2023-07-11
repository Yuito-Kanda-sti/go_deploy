package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"io"

	_ "github.com/go-sql-driver/mysql"
)

// `json:"id"`の記載がなくてもJSON形式で返却されるが、キー名のままだと困る場合はスネークケースなどに直すことができる
type Page struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

// Pageの配列リテラル（本来はDBから返却された値で埋めていく）
var pages = []Page{{
	ID:    1,
	Title: "The Go Programming Language",
	Url:   "https://golang.org/",
},
	{
		ID:    2,
		Title: "A Tour of Go",
		Url:   "https://go-tour-jp.appspot.com/welcome/1",
	},
}

// JSON返却用の構造体
type PageJSON struct {
	Status int `json:"status"`
	Pages  *[]Page
}

var (
	host     = os.Getenv("MYSQLCONNSTR_DBHOST")
	database = os.Getenv("MYSQLCONNSTR_DBDATABASENAME")
	user     = os.Getenv("MYSQLCONNSTR_DBUSER")
	password = os.Getenv("MYSQLCONNSTR_DBPASSWORD")
)

// ただただ文字列「Hello, World」を返却するハンドラー
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(os.Stdout,"this is standerd OUTPUT")
	fmt.Fprint(w, "Hello,World")

	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true&tls=true&timeout=10s", user, password, host, database)

	//db, err := sql.Open("mysql", user+":"+password+"@tcp("+host+":3306)/"+database+"?allowNativePasswords=true&tls=true")
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Fprint(w,err)
		fmt.Fprint(w, "Hello, Error ,World")
		return
	}
	err = db.Ping()
	if err != nil {
		fmt.Fprint(w,err)
		fmt.Fprint(w, "Hello, Error ,World")
		return
	}
	defer db.Close()
	var (
		id       int
		name     string
		quantity int
	)
	rows, err := db.Query("SELECT id, name, quantity from inventory;")
	if err != nil {
		fmt.Fprint(w,err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &name, &quantity)
		if err != nil {
			fmt.Fprint(w,err)
			return
		}
		fmt.Fprintf(w, "Data row = (%d, %s, %d)\n", id, name, quantity)
	}

	fmt.Fprint(w, "Hello, World")
	return
}

// appilication/jsonでJSONっぽい値を返却するハンドラー
func pagesHandler(w http.ResponseWriter, r *http.Request) {

	var pj PageJSON
	pj.Status = 200
	pj.Pages = &pages

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(&pj); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())

	// Content-Typeを設定
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	// Responseに書き込み
	_, err := fmt.Fprint(w, buf.String())
	if err != nil {
		return
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World")


	all, _ := ioutil.ReadDir("/storage")

	fileUrl := os.Getenv("TARGETIMAGE")
	fmt.Fprint(w, "image saving:")
	if err := DownloadFile("/storage/"+os.Getenv("TARGETIMAGENAME"), fileUrl); 
	err != nil {
        fmt.Fprint(w, " error:")
    }


	for _, f := range all {
		fmt.Fprint(w, f.Name()+" , ")
		// sample.go
		// test
		// test.txt
	}

	fmt.Fprint(w, "Hello, World")
	return

}

func DownloadFile(filepath string, url string) error {

    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}

func main() {
	fmt.Print("this is standerd OUTPUT on main()")

	// text/plain返却　"/" のとき indexHandlerを実行する
	http.HandleFunc("/", indexHandler)

	// application/json返却 "/pages"の時 pagesHandlerを返却
	http.HandleFunc("/pages", pagesHandler)

	http.HandleFunc("/save", saveHandler)

	log.Fatal(http.ListenAndServe(":80", nil))
}
