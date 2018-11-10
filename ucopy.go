package main

import (
	"flag"
	"fmt"
	"time"

	"simonwaldherr.de/go/golibs/csv"
	"simonwaldherr.de/go/saprfc"
)

// BAPI calls log
func printLog(bapi_return interface{}) {
	for _, line := range bapi_return.([]interface{}) {
		fmt.Printf("%s: %s\n",
			line.(map[string]interface{})["TYPE"],
			line.(map[string]interface{})["MESSAGE"])
	}
}

var sapConnPara *saprfc.Connection

func main() {

	// SAP Host connection settings
	var user string
	var pwd string
	var host string
	var router string
	var sysnr string
	var client string
	var trace string
	var lang string
	flag.StringVar(&user, "user", "username", "sap rfc username")
	flag.StringVar(&pwd, "pwd", "password", "sap rfc password")
	flag.StringVar(&host, "host", "10.11.12.13", "sap host")
	flag.StringVar(&router, "router", "/H/123.12.123.12/E/yt6ntx/H/123.14.131.111/H/", "sap router")
	flag.StringVar(&sysnr, "sysnr", "200", "system number")
	flag.StringVar(&client, "client", "300", "client")
	flag.StringVar(&trace, "trace", "3", "trace")
	flag.StringVar(&lang, "lang", "EN", "language")

	// csv file with input data
	var csvfile string
	flag.StringVar(&csvfile, "csv", "", "user data csv file")

	// csv file with input data
	var fromuser string
	flag.StringVar(&fromuser, "fromuser", "", "username of the source user")

	// load flag input arguments
	flag.Parse()

	sapConnPara := saprfc.ConnectionParameter{
		User:      user,
		Passwd:    pwd,
		Ashost:    host,
		Saprouter: router,
		Sysnr:     sysnr,
		Client:    client,
		Trace:     trace,
		Lang:      lang,
	}

	conn, _ := saprfc.ConnectionFromParams(sapConnPara)

	// The source user, to be copied
	unameFrom := "UNAMEFROM"

	// Defaults if source user validity not maintained (undefined)
	validFrom := time.Date(2018, time.January, 19, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)

	// Get source user details
	r, _ := conn.Call("BAPI_USER_GET_DETAIL", map[string]interface{}{"USERNAME": unameFrom, "CACHE_RESULTS": " "})

	// Set new users" defaults
	logonData := r["LOGONDATA"].(map[string]interface{})

	if logonData["GLTGV"] == nil {
		logonData["GLTGV"] = validFrom
	}
	if logonData["GLTGB"] == nil {
		logonData["GLTGB"] = validTo
	}

	// Create new users
	address := r["ADDRESS"].(map[string]interface{})

	csvmap, k := csv.LoadCSVfromFile(csvfile)

	for _, user := range csvmap {
		fmt.Println(user[k["username"]])

		address["LASTNAME"] = user[k["lastname"]]
		address["FULLNAME"] = user[k["firstname"]] + user[k["lastname"]]

		x, _ := conn.Call("BAPI_USER_CREATE1", map[string]interface{}{
			"USERNAME":  user[k["username"]],
			"LOGONDATA": logonData,
			"PASSWORD":  user[k["password"]],
			"DEFAULTS":  r["DEFAULTS"],
			"ADDRESS":   address,
			"COMPANY":   r["COMPANY"],
			"REF_USER":  r["REF_USER"],
			"PARAMETER": r["PARAMETER"],
			"GROUPS":    r["GROUPS"],
		})

		printLog(x["RETURN"])

		x, _ = conn.Call("BAPI_USER_PROFILES_ASSIGN", map[string]interface{}{
			"USERNAME": user[k["username"]],
			"PROFILES": r["PROFILES"],
		})

		printLog(x["RETURN"])

		x, _ = conn.Call("BAPI_USER_ACTGROUPS_ASSIGN", map[string]interface{}{
			"USERNAME":       user[k["username"]],
			"ACTIVITYGROUPS": r["ACTIVITYGROUPS"],
		})

		printLog(x["RETURN"])
	}

	// Finished
	fmt.Printf("%s copied to %d new users.\nBye!", unameFrom, len(csvmap))
}
