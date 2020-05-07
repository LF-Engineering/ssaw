package main

import (
	"database/sql"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsns "github.com/aws/aws-sdk-go/service/sns"
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

func getThreadsNum() int {
	nCPUsStr := os.Getenv("N_CPUS")
	nCPUs := 0
	if nCPUsStr != "" {
		var err error
		nCPUs, err = strconv.Atoi(nCPUsStr)
		if err != nil || nCPUs < 0 {
			nCPUs = 0
		}
	}
	if nCPUs > 0 {
		n := runtime.NumCPU()
		if nCPUs > n {
			nCPUs = n
		}
		runtime.GOMAXPROCS(nCPUs)
		return nCPUs
	}
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	return thrN
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

func processOrg(ch chan [3]string, apiURL, lfAuth, org string, updatedAt time.Time) (ret [3]string) {
	defer func() {
		if recover() != nil {
			mPrintf("org %s, updated at: %v error:\n%s\n", org, updatedAt, string(debug.Stack()))
		}
		if ch != nil {
			ch <- ret
		}
	}()
	method := "GET"
	xRequestID := fmt.Sprintf("sync-from-sfdc-%s{{%s}}", time.Now().Format(time.RFC3339Nano), org)
	params := url.Values{}
	params.Add("name", "["+org+"]")
	surl := fmt.Sprintf("%s/orgs/search?%s", apiURL, params.Encode())
	req, err := http.NewRequest(method, surl, nil)
	if err != nil {
		err = fmt.Errorf("new request error: %+v for %s url: %s\n", err, method, surl)
		fatalOnError(err, false)
		return
	}
	//req.Header.Set("X-ACL", lfAuth)
	ary := strings.Split(lfAuth, ":")
	req.Header.Set("X-ACL", ary[0])
	req.Header.Set("Bearer", ary[1])
	req.Header.Set("X-REQUEST-ID", xRequestID)
	mPrintf("request: %+v\n", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("do request error: %+v for %s url: %s\n", err, method, surl)
		fatalOnError(err, false)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("ReadAll non-ok request error: %+v for %s url: %s\n", err, method, surl)
			fatalOnError(err, false)
			return
		}
		err = fmt.Errorf("Method:%s url:%s status:%d\n%s\n", method, surl, resp.StatusCode, body)
		fatalOnError(err, false)
		return
	}
	/*
	  var payload allArrayOutput
	  err = yaml.NewDecoder(resp.Body).Decode(&payload)
	  if err != nil {
	    body, err2 := ioutil.ReadAll(resp.Body)
	    if err2 != nil {
	      err2 = fmt.Errorf("ReadAll yaml request error: %+v, %+v for %s url: %s\n", err, err2, method, surl)
	      fatalOnError(err, false)
	      return
	    }
	    err = fmt.Errorf("yaml decode error: %+v for %s url: %s\nBody: %s\n", err, method, surl, body)
	    fatalOnError(err, false)
	    return
	  }
	  ok = true
	  profs = payload.Profiles
	*/
	ret = [3]string{"org", org, ""}
	//mPrintf("org: %s, updated at: %v\n", org, updatedAt)
	return
}

func processProfile(ch chan [3]string, apiURL, lfAuth, uuid string, updatedAt time.Time) (ret [3]string) {
	defer func() {
		if recover() != nil {
			mPrintf("profile %s, updated at: %v error:\n%s\n", uuid, updatedAt, string(debug.Stack()))
		}
		if ch != nil {
			ch <- ret
		}
	}()
	//mPrintf("profile: %s, updated at: %v\n", uuid, updatedAt)
	ret = [3]string{"profile", uuid, ""}
	return
}

func processTopic(region, key, secret, topic string) {
	defer func() {
		if recover() != nil {
			mPrintf("%s\n", string(debug.Stack()))
		}
	}()
	sns := awsns.New(
		session.Must(
			session.NewSession(
				&aws.Config{
					Region: aws.String(region),
					// id, secret, token
					Credentials: credentials.NewStaticCredentials(key, secret, ""),
					MaxRetries:  aws.Int(5),
				},
			),
		),
	)
	mPrintf("%s: %+v\n", topic, sns)
	for {
		// FIXME: subscribe to SNS topic and fetch upadtes from it
		time.Sleep(10 * time.Second)
	}
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
	orgAPIURL := os.Getenv("ORG_SVC_URL")
	userAPIURL := os.Getenv("USER_SVC_URL")
	lfAuth := os.Getenv("LF_AUTH")
	thrN := getThreadsNum()
	mPrintf("Using %d CPUs\n", thrN)
	if thrN > 1 {
		ch := make(chan [3]string)
		nThreads := 0
		for index := range companies {
			go processOrg(ch, orgAPIURL, lfAuth, companies[index], modCompaniesAry[index])
			nThreads++
			if nThreads == thrN {
				res := <-ch
				mPrintf("finished %+v\n", res)
				nThreads--
			}
		}
		for index := range uuids {
			go processProfile(ch, userAPIURL, lfAuth, uuids[index], modUUIDsAry[index])
			nThreads++
			if nThreads == thrN {
				res := <-ch
				mPrintf("finished %+v\n", res)
				nThreads--
			}
		}
		for nThreads > 0 {
			res := <-ch
			mPrintf("finished %+v\n", res)
			nThreads--
		}
	} else {
		for index := range companies {
			processOrg(nil, orgAPIURL, lfAuth, companies[index], modCompaniesAry[index])
		}
		for index := range uuids {
			res := processProfile(nil, userAPIURL, lfAuth, uuids[index], modUUIDsAry[index])
			mPrintf("finished %+v\n", res)
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "SYNC_OK")
}

func handleSyncFromSFDC() {
	region := os.Getenv("AWS_REGION")
	key := os.Getenv("AWS_KEY")
	secret := os.Getenv("AWS_SECRET")
	topic := os.Getenv("AWS_TOPIC")
	for {
		processTopic(region, key, secret, topic)
		mPrintf("process topic finished, restarting\n")
	}
}

func initSHDB() {
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
	gDB = d
	_, err = exec(nil, "set @origin = ?", "exampleOrigin")
	if err != nil {
		fatalf(true, "unable to connect to origin session variable: %v", err)
	}
}

func checkEnv() {
	requiredEnv := []string{
		"SH_DB_ENDPOINT",
		"ORG_SVC_URL",
		"USER_SVC_URL",
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
	initSHDB()
	go handleSyncFromSFDC()
	http.HandleFunc("/sync-to-sfdc", handleSyncToSFDC)
	fatalOnError(http.ListenAndServe("0.0.0.0:6060", nil), true)
}

func main() {
	serve()
	fatalf(true, "serve exited without error, returning error state anyway")
}
