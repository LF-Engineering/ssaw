package main

import (
	"database/sql"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	dateTimeFormatMillis = "2006-01-02 15:04:05.999"
)

var (
	gMtx *sync.Mutex
	gw   http.ResponseWriter
	gDB  *sqlx.DB
)

func mPrintf(format string, args ...interface{}) (n int, err error) {
	now := time.Now()
	n, err = fmt.Printf("%s", fmt.Sprintf("%s: "+format, append([]interface{}{now.Format(dateTimeFormatMillis)}, args...)...))
	return
}

func timeStampStr() string {
	return time.Now().Format(dateTimeFormatMillis) + ": "
}

func fatalOnError(err error, pnic bool) bool {
	if err != nil {
		tm := time.Now()
		mPrintf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Fprintf(os.Stderr, "Error(time=%+v):\nError: '%s'\nStacktrace:\n", tm, err.Error())
		if gw != nil {
			gw.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(gw, timeStampStr()+err.Error()+"\n")
		}
		if pnic {
			panic("stacktrace")
		}
		return true
	}
	return false
}

func fatalf(pnic bool, f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...), pnic)
}

func queryOut(query string, args ...interface{}) {
	mPrintf("%s\n", query)
	if len(args) > 0 {
		s := ""
		for vi, vv := range args {
			switch v := vv.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, string, bool, time.Time:
				s += fmt.Sprintf("%d:%+v ", vi+1, v)
			case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *complex64, *complex128, *string, *bool, *time.Time:
				s += fmt.Sprintf("%d:%+v ", vi+1, v)
			case nil:
				s += fmt.Sprintf("%d:(null) ", vi+1)
			default:
				s += fmt.Sprintf("%d:%+v ", vi+1, reflect.ValueOf(vv))
			}
		}
		mPrintf("[%s]\n", s)
	}
}

func queryDB(query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = gDB.Query(query, args...)
	if err != nil {
		queryOut(query, args...)
	}
	return
}

func queryTX(tx *sql.Tx, query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = tx.Query(query, args...)
	if err != nil {
		queryOut(query, args...)
	}
	return
}

func query(tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tx == nil {
		return queryDB(query, args...)
	}
	return queryTX(tx, query, args...)
}

func execDB(query string, args ...interface{}) (res sql.Result, err error) {
	res, err = gDB.Exec(query, args...)
	if err != nil {
		queryOut(query, args...)
	}
	return
}

func execTX(tx *sql.Tx, query string, args ...interface{}) (res sql.Result, err error) {
	res, err = tx.Exec(query, args...)
	if err != nil {
		queryOut(query, args...)
	}
	return
}

func exec(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tx == nil {
		return execDB(query, args...)
	}
	return execTX(tx, query, args...)
}

func requestInfo(r *http.Request) string {
	agent := ""
	hdr := r.Header
	method := r.Method
	path := html.EscapeString(r.URL.Path)
	if hdr != nil {
		uAgentAry, ok := hdr["User-Agent"]
		if ok {
			agent = strings.Join(uAgentAry, ", ")
		}
	}
	if agent != "" {
		return fmt.Sprintf("IP: %s, agent: %s, method: %s, path: %s", r.RemoteAddr, agent, method, path)
	}
	return fmt.Sprintf("IP: %s, method: %s, path: %s", r.RemoteAddr, method, path)
}

func handleSyncToSFDC(w http.ResponseWriter, req *http.Request) {
	gw = w
	info := requestInfo(req)
	mPrintf("Request: %s\n", info)
	var err error
	defer func() {
		mPrintf("Request(exit): %s err:%v\n", info, err)
	}()
	mPrintf("lock mutex\n")
	gMtx.Lock()
	defer func() {
		mPrintf("unlock mutex\n")
		gMtx.Unlock()
	}()
	var (
		modified        time.Time
		modUUIDsAry     []time.Time
		modCompaniesAry []time.Time
		uuid            string
		uuids           []string
		company         string
		companies       []string
	)
	rows, err := query(nil, "select last_modified, name from orgs_for_sf_sync")
	if fatalOnError(err, false) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&modified, &company)
		if fatalOnError(err, false) {
			return
		}
		modCompaniesAry = append(modCompaniesAry, modified)
		companies = append(companies, company)
	}
	err = rows.Err()
	if fatalOnError(err, false) {
		return
	}
	err = rows.Close()
	if fatalOnError(err, false) {
		return
	}
	rows, err = query(nil, "select last_modified, uuid from uuids_for_sf_sync")
	if fatalOnError(err, false) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&modified, &uuid)
		if fatalOnError(err, false) {
			return
		}
		modUUIDsAry = append(modUUIDsAry, modified)
		uuids = append(uuids, uuid)
	}
	err = rows.Err()
	if fatalOnError(err, false) {
		return
	}
	err = rows.Close()
	if fatalOnError(err, false) {
		return
	}
	mPrintf("%d companies to process: %+v\n", len(companies), companies)
	mPrintf("%d UUIDs to process: %+v\n", len(uuids), uuids)
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "SYNC_OK")
}

func initSHDB() *sqlx.DB {
	dbURL := os.Getenv("SH_DB_ENDPOINT")
	if !strings.Contains(dbURL, "parseTime=true") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&parseTime=true"
		} else {
			dbURL += "?parseTime=true"
		}
	}
	d, err := sqlx.Connect("mysql", dbURL)
	if err != nil {
		fatalf(true, "unable to connect to affiliation database: %v", err)
	}
	d.SetConnMaxLifetime(time.Second)
	return d
}

func checkEnv() {
	requiredEnv := []string{
		"SH_DB_ENDPOINT",
		"AWS_REGION",
		"AWS_KEY",
		"AWS_SECRET",
		"AWS_TOPIC",
		"LF_AUTH",
	}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			fatalf(true, "%s env variable must be set", env)
		}
	}
}

func serve() {
	mPrintf("Starting sync server\n")
	checkEnv()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGALRM)
	go func() {
		for {
			sig := <-sigs
			mPrintf("Exiting due to signal %v\n", sig)
			os.Exit(1)
		}
	}()
	gMtx = &sync.Mutex{}
	gDB = initSHDB()
	http.HandleFunc("/sync-to-sfdc", handleSyncToSFDC)
	fatalOnError(http.ListenAndServe("0.0.0.0:6060", nil), true)
}

func main() {
	serve()
	fatalf(true, "serve exited without error, returning error state anyway")
}
