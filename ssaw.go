package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

	"github.com/LF-Engineering/ssaw/ssawsync"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsns "github.com/aws/aws-sdk-go/service/sns"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	dateTimeFormatMillis = "2006-01-02 15:04:05.999"
	cAll                 = "all"
)

var (
	gMtx               *sync.Mutex
	gTokenMtx          *sync.Mutex
	gw                 http.ResponseWriter
	gDB                *sqlx.DB
	gAuth0URL          string
	gAuth0ClientID     string
	gAuth0ClientSecret string
	gAuth0Audience     string
	gLFAuth            string
	gUserAPIURL        string
	gOrgAPIURL         string
	gAffAPIURL         string
	gNotifAPIURL       string
	gGitdmURL          string
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

func jsonEscape(str string) string {
	b, _ := json.Marshal(str)
	return string(b[1 : len(b)-1])
}

func getToken() (err error) {
	defer func() {
		if recover() != nil {
			mPrintf("getToken eror:\n%s\n", err, string(debug.Stack()))
		}
	}()
	data := fmt.Sprintf(
		`{"grant_type":"client_credentials","client_id":"%s","client_secret":"%s","audience":"%s","scope":"access:api"}`,
		jsonEscape(gAuth0ClientID),
		jsonEscape(gAuth0ClientSecret),
		jsonEscape(gAuth0Audience),
	)
	payloadBytes := []byte(data)
	payloadBody := bytes.NewReader(payloadBytes)
	method := http.MethodPost
	surl := fmt.Sprintf("%s/oauth/token", gAuth0URL)
	req, e := http.NewRequest(method, surl, payloadBody)
	if e != nil {
		err = fmt.Errorf("new request error: %+v for %s url: %s\n", e, method, surl)
		fatalOnError(err, false)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		err = fmt.Errorf("do request error: %+v for %s url: %s\n", e, method, surl)
		fatalOnError(err, false)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			err = fmt.Errorf("ReadAll non-ok request error: %+v for %s url: %s\n", err, method, surl)
			fatalOnError(err, false)
			return
		}
		err = fmt.Errorf("Method:%s url:%s status:%d\n%s\n", method, surl, resp.StatusCode, body)
		fatalOnError(err, false)
		return
	}
	var rdata struct {
		Token string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&rdata)
	if err != nil {
		fatalOnError(err, false)
		return
	}
	if rdata.Token == "" {
		fatalOnError(fmt.Errorf("empty token retuned"), false)
		return
	}
	gLFAuth = "Bearer " + rdata.Token
	mPrintf("Received new token %s (length %d)\n", gLFAuth, len(gLFAuth))
	return
}

/*
func processOrg(ch chan [3]string, org string, updatedAt time.Time, src, op string) (ret [3]string) {
	var err error
	defer func() {
		if recover() != nil {
			mPrintf("org %s, updated at %v, src %s, op %s, error:\n%s\n", org, updatedAt, src, op, string(debug.Stack()))
		}
		if err != nil {
			ret[2] = err.Error()
		}
		if ch != nil {
			ch <- ret
		}
	}()
	for i := 0; i < 2; i++ {
		method := http.MethodGet
		params := url.Values{}
		params.Add("name", org)
		surl := fmt.Sprintf("%s/orgs/search?%s", gOrgAPIURL, params.Encode())
		req, e := http.NewRequest(method, surl, nil)
		if e != nil {
			err = fmt.Errorf("new request error: %+v for %s url: %s\n", e, method, surl)
			fatalOnError(err, false)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", gLFAuth)
		//mPrintf("request: %+v\n", req)
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			err = fmt.Errorf("do request error: %+v for %s url: %s\n", e, method, surl)
			fatalOnError(err, false)
			return
		}
		if i == 0 && resp.StatusCode == 401 {
			currToken := gLFAuth
			_ = resp.Body.Close()
			mPrintf("Token is invalid, trying to generate another one\n")
			gTokenMtx.Lock()
			if currToken == gLFAuth {
			  mPrintf("Generating new token\n")
				err = getToken()
			}
			gTokenMtx.Unlock()
			if err != nil {
				fatalOnError(err, false)
				return
			}
			continue
		}
		if resp.StatusCode != 200 {
			body, e := ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if e != nil {
				err = fmt.Errorf("ReadAll non-ok request error: %+v for %s url: %s\n", e, method, surl)
				fatalOnError(err, false)
				return
			}
			err = fmt.Errorf("Method:%s url:%s status:%d\n%s\n", method, surl, resp.StatusCode, body)
			fatalOnError(err, false)
			return
		}
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			err = e
			fatalOnError(err, false)
			return
		}
		_ = resp.Body.Close()
		mPrintf("%s\n", body)
		ret = [3]string{"org", org, ""}
		break
	}
	return
}
*/

func regenerateToken() (err error) {
	currToken := gLFAuth
	gTokenMtx.Lock()
	if currToken == gLFAuth {
		mPrintf("Generating new token\n")
		err = getToken()
	}
	gTokenMtx.Unlock()
	return
}

func processOrg(ch chan [3]string, org string, updatedAt time.Time, src, op string) (ret [3]string) {
	mPrintf("processOrg: %s\n", org)
	var err error
	defer func() {
		if recover() != nil {
			mPrintf("org %s, updated at %v, src %s, op %s, error:\n%s\n", org, updatedAt, src, op, string(debug.Stack()))
		}
		if err != nil {
			ret[2] = err.Error()
		}
		if ch != nil {
			ch <- ret
		}
	}()
	for i := 0; i < 2; i++ {
		method := http.MethodGet
		params := url.Values{}
		params.Add("name", org)
		surl := fmt.Sprintf("%s/orgs/search?%s", gOrgAPIURL, params.Encode())
		req, e := http.NewRequest(method, surl, nil)
		if e != nil {
			err = fmt.Errorf("new request error: %+v for %s url: %s\n", e, method, surl)
			fatalOnError(err, false)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", gLFAuth)
		//mPrintf("request: %+v\n", req)
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			err = fmt.Errorf("do request error: %+v for %s url: %s\n", e, method, surl)
			fatalOnError(err, false)
			return
		}
		if i == 0 && resp.StatusCode == 401 {
			_ = resp.Body.Close()
			mPrintf("Token is invalid, trying to generate another one\n")
			err = regenerateToken()
			if err != nil {
				fatalOnError(err, false)
				return
			}
			continue
		}
		if resp.StatusCode != 200 {
			body, e := ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if e != nil {
				err = fmt.Errorf("ReadAll non-ok request error: %+v for %s url: %s\n", e, method, surl)
				fatalOnError(err, false)
				return
			}
			err = fmt.Errorf("Method:%s url:%s status:%d\n%s\n", method, surl, resp.StatusCode, body)
			fatalOnError(err, false)
			return
		}
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			err = e
			fatalOnError(err, false)
			return
		}
		_ = resp.Body.Close()
		mPrintf("%s\n", body)
		ret = [3]string{"org", org, ""}
		break
	}
	return
}

func processProfile(ch chan [3]string, uuid string, updatedAt time.Time, src, op string) (ret [3]string) {
	var err error
	defer func() {
		if recover() != nil {
			mPrintf("profile %s, updated at %v, src %s, op %s, error:\n%s\n", uuid, updatedAt, src, op, string(debug.Stack()))
		}
		if err != nil {
			ret[2] = err.Error()
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
	mPrintf("%+v\n", sns)
	for {
		// FIXME: subscribe to SNS topic and fetch updates from it
		time.Sleep(10 * time.Second)
	}
}

// This is called from: ssawsync/sync.go (ssawsync.Sync)
func sendToSNS(w http.ResponseWriter, req *http.Request) {
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
	path := html.EscapeString(req.URL.Path)
	// /sync/origin-name
	// for example: /sync/json2hat
	ary := strings.Split(path, "/")
	if len(ary) != 3 {
		fatalf(false, "malformed path:%s", path)
		return
	}
	origin := ary[2]
	mPrintf("origin: %s\n", origin)
	if gGitdmURL == "" {
		gGitdmURL = os.Getenv("GITDM_SYNC_URL")
	}
	switch origin {
	case "json2hat", "bitergia-import-sh-json", "bitergia-import-map-file", "sds-final":
		e := ssawsync.SyncGitdm(gGitdmURL, origin)
		if e != nil {
			mPrintf("gitdm sync error for %s origin: %v\n", origin, e)
		}
	case "da-affiliation-api", "gitdm", "sds-partial":
		mPrintf("Not calling gitdm sync for origin %s\n", origin)
	default:
		mPrintf("unknown origin: %s - not calling gitdm sync\n", origin)
	}
	gTokenMtx = &sync.Mutex{}
	if gAuth0URL == "" {
		gAuth0URL = os.Getenv("AUTH0_URL")
	}
	if gAuth0ClientID == "" {
		gAuth0ClientID = os.Getenv("AUTH0_CLIENT_ID")
	}
	if gAuth0ClientSecret == "" {
		gAuth0ClientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	}
	if gAuth0Audience == "" {
		gAuth0Audience = os.Getenv("AUTH0_AUDIENCE")
	}
	if gNotifAPIURL == "" {
		gNotifAPIURL = os.Getenv("NOTIF_SVC_URL")
	}
	if gOrgAPIURL == "" {
		gOrgAPIURL = os.Getenv("ORG_SVC_URL")
	}
	if gUserAPIURL == "" {
		gUserAPIURL = os.Getenv("USER_SVC_URL")
	}
	if gAffAPIURL == "" {
		gAffAPIURL = os.Getenv("AFF_API_URL")
	}
	// can be used during development, specify last received token
	// to avoid getting new token every time app is restarted
	gLFAuth = os.Getenv("BEARER_TOKEN")
	if gLFAuth == "" {
		err = getToken()
		if fatalOnError(err, false) {
			return
		}
	}
	thrN := getThreadsNum()
	var (
		company   string
		uuid      string
		modified  time.Time
		src       string
		op        string
		companies []string
		uuids     []string
		modifieds []time.Time
		srcs      []string
		ops       []string
		rows      *sql.Rows
	)
	// organizations
	// rows, err = query(nil, "select name, last_modified, src, op from sync_orgs where src = ?", origin)
	// FIXME
	rows, err = query(nil, "select name, max(last_modified) from sync_orgs group by name order by name limit 4")
	if fatalOnError(err, false) {
		return
	}
	for rows.Next() {
		//err = rows.Scan(&company, &modified, &src, &op)
		err = rows.Scan(&company, &modified)
		if fatalOnError(err, false) {
			return
		}
		companies = append(companies, company)
		modifieds = append(modifieds, modified)
		//srcs = append(srcs, src)
		//ops = append(ops, op)
	}
	err = rows.Err()
	if fatalOnError(err, false) {
		return
	}
	err = rows.Close()
	if fatalOnError(err, false) {
		return
	}
	mPrintf("%d companies to process\n", len(companies))
	mPrintf("Using %d CPUs\n", thrN)
	failedOrgs := 0
	if thrN > 1 {
		ch := make(chan [3]string)
		nThreads := 0
		for index := range companies {
			//go processOrg(ch, companies[index], modifieds[index], srcs[index], ops[index])
			go processOrg(ch, companies[index], modifieds[index], "", "")
			nThreads++
			if nThreads == thrN {
				res := <-ch
				mPrintf("finished %+v\n", res)
				nThreads--
				if res[2] != "" {
					failedOrgs++
				}
			}
		}
		for nThreads > 0 {
			res := <-ch
			mPrintf("finished %+v\n", res)
			nThreads--
			if res[2] != "" {
				failedOrgs++
			}
		}
	} else {
		for index := range companies {
			//res := processOrg(nil, companies[index], modifieds[index], srcs[index], ops[index])
			res := processOrg(nil, companies[index], modifieds[index], "", "")
			mPrintf("finished %+v\n", res)
			if res[2] != "" {
				failedOrgs++
			}
		}
	}

	// profiles
	modifieds = []time.Time{}
	srcs = []string{}
	ops = []string{}
	//rows, err = query(nil, "select uuid, last_modified, src, op from sync_uuids where src = ?", origin)
	// FIXME
	rows, err = query(nil, "select uuid, last_modified, src, op from sync_uuids limit 1")
	if fatalOnError(err, false) {
		return
	}
	for rows.Next() {
		err = rows.Scan(&uuid, &modified, &src, &op)
		if fatalOnError(err, false) {
			return
		}
		uuids = append(uuids, uuid)
		modifieds = append(modifieds, modified)
		srcs = append(srcs, src)
		ops = append(ops, op)
	}
	err = rows.Err()
	if fatalOnError(err, false) {
		return
	}
	err = rows.Close()
	if fatalOnError(err, false) {
		return
	}
	mPrintf("%d UUIDs to process\n", len(uuids))
	mPrintf("Using %d CPUs\n", thrN)
	failedProfiles := 0
	if thrN > 1 {
		ch := make(chan [3]string)
		nThreads := 0
		for index := range uuids {
			go processProfile(ch, uuids[index], modifieds[index], srcs[index], ops[index])
			nThreads++
			if nThreads == thrN {
				res := <-ch
				mPrintf("finished %+v\n", res)
				nThreads--
				if res[2] != "" {
					failedProfiles++
				}
			}
		}
		for nThreads > 0 {
			res := <-ch
			mPrintf("finished %+v\n", res)
			nThreads--
			if res[2] != "" {
				failedProfiles++
			}
		}
	} else {
		for index := range uuids {
			res := processProfile(nil, uuids[index], modifieds[index], srcs[index], ops[index])
			mPrintf("finished %+v\n", res)
			if res[2] != "" {
				failedProfiles++
			}
		}
	}
	if failedOrgs > 0 || failedProfiles > 0 {
		fatalf(false, "failed organizations:%d profiles:%d", failedOrgs, failedProfiles)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "SYNC_OK")
}

func subscribeToSNS() {
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
	d.SetConnMaxLifetime(30 * time.Second)
	gDB = d
	_, err = exec(nil, "set @origin = ?", "sfdc")
	if err != nil {
		fatalf(true, "unable to set origin session variable: %v", err)
	}
}

func checkEnv() {
	requiredEnv := []string{
		"SH_DB_ENDPOINT",
		"GITDM_SYNC_URL",
		"NOTIF_SVC_URL",
		"ORG_SVC_URL",
		"USER_SVC_URL",
		"AFF_API_URL",
		"AWS_REGION",
		"AWS_KEY",
		"AWS_SECRET",
		"AWS_TOPIC",
		"AUTH0_URL",
		"AUTH0_AUDIENCE",
		"AUTH0_CLIENT_ID",
		"AUTH0_CLIENT_SECRET",
	}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			fatalf(true, "%s env variable must be set", env)
		}
	}
}

func serve() {
	mPrintf("Starting serve\n")
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
	go subscribeToSNS()
	http.HandleFunc("/sync/", sendToSNS)
	fatalOnError(http.ListenAndServe("0.0.0.0:6060", nil), true)
}

func main() {
	serve()
	fatalf(true, "serve exited without error, returning error state anyway")
}
