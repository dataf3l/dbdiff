package main

/**
The purpose of this program is to allow the easy comparison of
2 different tables in a single database,
This can be useful since we want to compare different databases
it supports MySQL for now, but support for more databases can be added
in the future
it supports tables, but support can be added for schema comparison,
record count basic reporting comparison, code diffing, etc.

This program can probably be used for people who have a "development"
database and a "production" database so they can figure out the
differences between development and production

We sincerely hope this program will be useful at some point.

*/

import (
	"database/sql"
	"fmt"
	"regexp"
	//"io"
	//"io/ioutil"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	//"time"
	"encoding/json"
	"github.com/robertkrimen/otto"
	"io/ioutil"
)

var db1name string
var db2name string
var configFile string
var changes []string

// getDatabaseConnection provides a database connection *sql.DB Object
// https://golang.org/pkg/database/sql/#DB
func getDatabaseConnection(dbName string) *sql.DB {
	connectionString := getConnectionString(dbName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}

// TODO: remove
func checkErr(e error) {
	if e == nil {
		panic(e)
	}
}

// retrieves the connection from the database
// TODO: add PORT
func getConnectionString(dbName string) string {
	/* constants which contains information about mysql database connectivity */
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")
	//dbname := os.Getenv("DBNAME")
	dbhost := os.Getenv("DBHOST")
	dbport := os.Getenv("DBPORT")

	connectionString := dbuser + ":" + dbpass + "@tcp" + "(" + dbhost + ":" + dbport + ")/" + dbName
	if dbuser == "" {
		log.Println("BAD ENVIRONMENT, DBUSER NOT SET.")
	}
	return connectionString
}

// Site represents a record on the Site table, although it
// could probably be any table, our table is called Site, perhaps
// rename it to Record?
type Site map[string]string

// getHeaders() creates the headers in HTML,
// TODO: Freeze headers, improve the width.
// There are 2 rows of headers, one "main header row" which is split
// into two subheaders, one subheader per database name, like this:
//
// The output of the program should kinda look like this:
// +----------------------------+
// |                            |
// |  field name goes here      |
// |                            |
// +-------------+--------------+
// |             |              |
// |  db1        |   db2        |
// |             |              |
// +-------------+--------------+
//
// Todo: Improve this, maybe add SQL Comment field below the header, as tooltip?

func getHeaders(db1name string, db2name string) string {
	dx := "<tr>"
	for _, h := range getFieldList() {
		dx += "<th colspan=3 class='field-name'>" + h + "</th>"
	}
	dx += "</tr>"
	dx += "<tr>"
	for range getFieldList() { // _, _ =
		dx += "<th>" + db1name + "</th>"
		dx += "<th style='width:5px'>&nbsp;</th>"
		dx += "<th>" + db2name + "</th>"
	}
	dx += "</tr>"

	return dx

}

// mkEmpty creates empty records so they can later be filled by the
// other function
// todo: OrderedMap
func mkEmpty() Site {
	d := make(Site)
	for _, f := range getFieldList() {
		d[f] = "EMPTY"
	}
	return d
}

// getFieldList returns a list of fields
// todo: get these fields directly from the query
// by parsing the query
// downside: no select *
//
// TODO: we can probably get the fields directly from the database
// using the schema tables (although, not portably)
func getFieldList() []string {
	/*
		fields := []string{
			"id",
			//	"name",
			//	"site",
			//	"site_url",
			//	"company_name",
			"job_title",
			//"job_description",
			"educationRequirements",
			"experienceRequirements",
			"qualifications",
			"responsibilities",
			"skills",
			"value_hour",
			//"sid",
			//"folder",
			//"offer_modulus",
			"enabled",
			"destination",
			"organization",
			"occupational_category",
			//"organization_logo",	// fix at the end organization_logo
			"script_template"}
	*/
	fields := []string{
		"id",
		"name",
		"site",
		"site_url",
		"company_name",
		"job_title",
		"job_description",
		"educationRequirements",
		"experienceRequirements",
		"qualifications",
		"responsibilities",
		"skills",
		"value_hour",
		"sid",
		"folder",
		"offer_modulus",
		"enabled",
		"destination",
		"organization",
		"occupational_category",
		"organization_logo", // fix at the end organization_logo
		"script_template"}

	return fields
}

// TODO: make a function map so people can indicate which
// columns should be different and which columns should be the same
// also exceldiff?????
func compareRecords(vm *otto.Otto, site1 Site, site2 Site, database1Name string, database2Name string, idx int) (string, int) {
	sidx := fmt.Sprintf("%d", idx)
	fields := getFieldList()
	diffCount := 0
	var resultFields []string
	result := "<tr>"

	//leftValues := []string{}
	//for _, fieldName := range fields {
	//	leftValue := site1[fieldName]
	//	leftValues = append(leftValues, fmt.Sprintf("'%s'",leftValue))
	//}

	//leftInsert := "INSERT INTO site (" + strings.Join(getFields(),",")+") VALUES('" + strings.Join(leftValues,",")+"');"

	//for _, fieldName := range fields {

	// figure out if everything is OK:

	rowIsIdentical := true
	for _, fieldName := range fields {
		leftValue := site1[fieldName]
		rightValue := site2[fieldName]

		if leftValue != rightValue {
			rowIsIdentical = false
			break
		}
	}

	for _, fieldName := range fields {
		leftValue := site1[fieldName]
		rightValue := site2[fieldName]
		lcss := getRecordCSS(vm, db1name, site1, fieldName)
		rcss := getRecordCSS(vm, db2name, site2, fieldName)

		lv := "<xmp>" + leftValue + "</xmp>"
		rv := "<xmp>" + rightValue + "</xmp>"

		// add rules
		okClassName := "ok"
		if rowIsIdentical {
			okClassName = "ok identical "
		}
		upref := "<td class='ucell " + okClassName + "' title='" + sidx + "'   >" // onclick='prompt(\"\",\"" + sidx + "\");'
		//upref = "<td class='ucell'>"

		lres := "<td class='lcell " + okClassName + "' style='" + lcss + "'>" + lv + "</td>"
		ures := upref + "</td>"
		rres := "<td class='rcell " + okClassName + "' style='" + rcss + "'>" + rv + "</td>"

		if leftValue == "EMPTY" {
			lres = "<td class='lcell empty'>&nbsp;</td>"
			ures = upref + "<a onclick='addChange(\"INS\",\"R2L\",\"ALL\",\"" + sidx + "\",event,this);'>&lt;</a></td>"
			// TODO: add another arrow for DELETION on L2R
			rres = "<td class='rcell' style='" + rcss + "'>" + rv + "</td>"
			diffCount++
			resultFields = append(resultFields, lres+ures+rres)
			continue
		}
		if rightValue == "EMPTY" {
			lres = "<td class='lcell' style='" + lcss + "'>" + lv + "</td>"
			ures = upref + "<a onclick='addChange(\"INS\",\"L2R\",\"ALL\",\"" + sidx + "\",event,this);'>&gt;</a></td>"
			// TODO: add another arrow for DELETION on R2L
			rres = "<td class='rcell empty'>&nbsp;</td>"
			diffCount++
			resultFields = append(resultFields, lres+ures+rres)
			continue
		}

		if leftValue == rightValue {
			// Ignore differences
		} else {
			// Report Differences
			diffCount++
			lres = "<td class='lcell' style='background-color:rgb(230,200,200);" + lcss + "'>" + lv + "</td>"
			ures = upref

			ures += "<a onclick='addChange(\"UPD\",\"R2L\",\"" + fieldName + "\",\"" + sidx + "\",event,this);'>&lt;</a><br/>"
			ures += "<br/>"
			ures += " <a onclick='addChange(\"UPD\",\"L2R\",\"" + fieldName + "\",\"" + sidx + "\",event,this);'>&gt;</a></td>"

			rres = "<td class='rcell' style='background-color:rgb(230,200,200)' style='" + rcss + "'>" + rv + "</td>"
		}

		resultFields = append(resultFields, lres+ures+rres)
	}
	result += strings.Join(resultFields, "")
	result += "</tr>\n"
	return result, diffCount
}

// used for generic database user interface
// apologies for the lack of types.
func makeResultReceiver(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{}
		current = struct{}{}
		result = append(result, &current)
	}
	return result
} // https://github.com/jinzhu/gorm/issues/1167

// rows, ids
func getRecords(query string, dbName string) (map[string]Site, []int, error) {
	ids := []int{}
	db := getDatabaseConnection(dbName)

	dbrows, err := db.Query(query)
	if err != nil {
		msg := fmt.Sprintf("fail : %s", err.Error())
		fmt.Println(msg)
		//io.WriteString(w, "Fail")
		db.Close()
		return make(map[string]Site), ids, err
	}
	result := make(map[string]Site)
	length := len(getFieldList())
	columns := getFieldList()
	for dbrows.Next() {

		current := makeResultReceiver(length)
		if err := dbrows.Scan(current...); err != nil {
			panic(err)
		}
		record := make(map[string]string)
		for i := 0; i < length; i++ {
			k := columns[i]
			//v := current[i]
			//record[k] = v.(string) // bad bad bad

			val := *(current[i]).(*interface{})
			if val == nil {
				record[k] = "NULL"
				continue
			}
			vType := reflect.TypeOf(val)
			switch vType.String() {
			case "int64":
				record[k] = fmt.Sprintf("%d", val.(int64))
			case "string":
				record[k] = val.(string)
			case "time.Time":
				record[k] = "DATDAT" // val.(time.Time)
			case "[]uint8":
				record[k] = string(val.([]uint8))
			default:
				fmt.Printf("unsupport data type '%s' now\n", vType)
				// TODO remember add other data type
			}

		}
		idAsInt, err := strconv.Atoi(record["id"])
		if err != nil {
			fmt.Println("INVALID ID DATA:'" + record["id"] + "' not convertible to int.")
		}
		ids = append(ids, idAsInt)
		result[record["id"]] = record

		//io.WriteString(w, tmpl)
	}

	db.Close()
	return result, ids, nil
}

// given 2 lists of ids, it combines both lists, sorted.
// perhaps just appending, sorting and later de-duplicating
// would have been easier
//
// TODO: change algo to append, sort, dedup
// (in order to remove the nasty for loop)
func getCombined(ids1 []int, ids2 []int) []int {
	// start:
	res := []int{}
	lid := 0
	rid := 0
	lmax := len(ids1)
	rmax := len(ids2)
	for {
		smallest := -1
		if rid >= rmax && lid >= lmax {
			break
		}

		if lid >= lmax {
			smallest = ids2[rid]
			rid++
			res = append(res, smallest)
			continue
		}
		if rid >= rmax {
			smallest = ids1[lid]
			lid++
			res = append(res, smallest)
			continue
		}

		lval := ids1[lid]
		rval := ids2[rid]

		if lval == rval {
			smallest = ids1[lid]
			lid++
			rid++
			res = append(res, smallest)
			continue
		}
		if lval < rval {
			smallest = ids1[lid]
			lid++
			res = append(res, smallest)
			continue
		}
		if lval > rval {
			smallest = ids2[rid]
			rid++
			res = append(res, smallest)
			continue
		}

		res = append(res, smallest)
		if rid >= rmax && lid >= lmax {
			break
		}
	}
	return res

}
func uiMenu() string {
	return fmt.Sprintf(`
	<ul>
	<li><a href='/list'>List</a></li>
	<li><a href='/changes'>Show Current Change Set</a><br/><br/><br/></li>
	<li><a href='/clear'>Delete All Changes(%d)</a></li>
	</ul>
	`, len(changes))

}
func jpre(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
func pre(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")

	fmt.Fprintf(w, `<link rel='stylesheet' type='text/css' href='/css/style.css'> `)
	fmt.Fprintf(w, `<script src='/js/db_diff.js'></script>`)

}
func showDifferencesHandler(w http.ResponseWriter, r *http.Request) {
	pre(w)
	fmt.Println("/list")

	vm := getJSVM()

	// some fields
	//q1 := "SELECT id-200 as id, job_title, educationRequirements, experienceRequirements, qualifications, responsibilities, skills, value_hour, enabled, destination, organization, occupational_category, script_template FROM site WHERE id>=214 and id <=253  order by id "

	// all fields:
	//q1 := "SELECT id-200 as id , name, site, site_url, company_name, job_title, job_description, educationRequirements, experienceRequirements, qualifications, responsibilities, skills, value_hour, sid, folder, offer_modulus, enabled, destination, organization, occupational_category, organization_logo, script_template FROM site WHERE id>=214 and id <=253  order by id "
	//q2 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site WHERE id>=14 and id<=53 ORDER BY id "

	/// all fields all records
	//////q1 := "SELECT id as id, name, site, site_url, company_name, job_title, job_description, educationRequirements, experienceRequirements, qualifications, responsibilities, skills, value_hour, sid, folder, offer_modulus, enabled, destination, organization, occupational_category, organization_logo, script_template FROM site WHERE site='tutree.com' and id >= 1053 and id <= 1089 order by id "
	//////q2 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site WHERE  id >= 1053 and id <= 1089 ORDER BY id "

	//q1 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site WHERE site='tutree.com' and id>12 and id < 214 and id < 1053 ORDER BY id"
	//q2 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site WHERE id > 53 and id < 1001 and id < 4000 ORDER BY id "

	//q1 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site WHERE site='tutree.com' and id >=214 and id <=253  ORDER BY id"

	q1 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site ORDER BY id"
	q2 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site ORDER BY id "

	var leftRecordSet map[string]Site
	var rightRecordSet map[string]Site

	leftRecordSet, ids1, err := getRecords(q1, db1name)
	if err != nil {
		log.Fatal(err)
	}
	rightRecordSet, ids2, err := getRecords(q2, db2name)
	if err != nil {
		log.Fatal(err)
	}
	// all ids:
	combined := getCombined(ids1, ids2)

	html := uiMenu()
	html += "<table border=1 cellspacing=0 cellpadding=4>"
	html += getHeaders(db1name, db2name)
	total := 0
	for _, idxAsInt := range combined {
		idx := fmt.Sprintf("%d", idxAsInt)

		leftRecord := mkEmpty()
		rightRecord := mkEmpty()

		if val, ok := leftRecordSet[idx]; ok {
			leftRecord = val
		}

		if val, ok := rightRecordSet[idx]; ok {
			rightRecord = val
		}

		r, c := compareRecords(vm, leftRecord, rightRecord, db1name, db2name, idxAsInt)
		total += c
		html += r
	}
	html += "</table>"
	html += fmt.Sprintf("Total: %d", total)
	fmt.Fprintf(w, html)
}
func clearChangesHandler(w http.ResponseWriter, r *http.Request) {
	pre(w)
	changes = []string{}
	fmt.Fprintf(w, uiMenu()+"<br/><script>;alert(\"Changes were cleared!\");location.href='/list';</script>")
}
func listChangesHandler(w http.ResponseWriter, r *http.Request) {
	pre(w)
	cx := uiMenu()
	cx += "<h1>Changes</h1><pre>"
	for _, change := range changes {
		cx += change + "\n"
	}
	cx += "</pre>"
	fmt.Fprintf(w, cx)
}

// Eliminates bad stuff from input, we can't trust the user's data.
// it remove bad characters
func removeStrangeCharacters(a string) string {
	var re = regexp.MustCompile(`([\\"'<>&%$])`)
	return re.ReplaceAllString(a, "")
}

func mkInsert(row Site, dbName string) string {
	fields := []string{}
	values := []string{}
	for _, fieldName := range getFieldList() {
		fields = append(fields, fieldName)
		values = append(values, row[fieldName])
	}
	return "INSERT INTO " + dbName + ".site (" + strings.Join(fields, ",") + ") VALUES('" + strings.Join(values, "','") + "');"
}
func mkSingleFieldUpdate(row Site, dbName string, fieldName string) string {
	value := row[fieldName]
	return "UPDATE  " + dbName + ".site SET " + fieldName + " = '" + value + "' WHERE id = '" + row["id"] + "';"
}

// will produce insert, update, whatever, on demand.
func addChangeHandler(w http.ResponseWriter, r *http.Request) {
	jpre(w)
	r.ParseForm()

	if len(r.Form["changeType"]) == 0 || len(r.Form["direction"]) == 0 || len(r.Form["fieldName"]) == 0 || len(r.Form["id"]) == 0 {
		fmt.Fprintf(w, `{"result":"MISSING PARAMS"}`)
	}

	changeType := removeStrangeCharacters(r.Form["changeType"][0]) // INS or UPD
	direction := removeStrangeCharacters(r.Form["direction"][0])
	fieldName := removeStrangeCharacters(r.Form["fieldName"][0])
	id := removeStrangeCharacters(r.Form["id"][0]) // maybe comma separated
	fmt.Printf("append:%s:%s:%s:%s\n", changeType, direction, fieldName, id)

	idList := strings.Split(id, ",")

	dataSource := db1name
	destination := db2name
	if direction == "R2L" {
		dataSource = db2name
		destination = db1name
	}
	//get records:
	sqlWhereStatement := "1=2"
	if len(idList) == 1 {
		sqlWhereStatement = " (id = '" + id + "') "
	} else if len(idList) == 2 {
		sqlWhereStatement = " (id >= '" + idList[0] + "' and id <= '" + idList[1] + "') "
	}
	q := "SELECT * FROM site WHERE " + sqlWhereStatement
	rows, _, err := getRecords(q, dataSource)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, `{"result":"SQL FAIL"}`)
		return
	}
	jsonOutput := `{"result":"USE UPD OR INS"}`
	if changeType == "INS" {
		for _, row := range rows {
			changes = append(changes, mkInsert(row, destination))
		}
		jsonOutput = fmt.Sprintf(`{"result":"Inserted:%d to %s"}`, len(rows), destination)
	} else if changeType == "UPD" {
		for _, row := range rows {
			changes = append(changes, mkSingleFieldUpdate(row, destination, fieldName))
		}
		jsonOutput = fmt.Sprintf(`{"result":"Updated:%d to %s"}`, len(rows), destination)

		// TODO: DEL
	}
	fmt.Fprintf(w, jsonOutput)

}

func runConfigInJSVM(vm *otto.Otto) {
	//load configuration file:
	file, err := os.Open(configFile) // unlisted/config.js
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	configFileContents, err := ioutil.ReadAll(file)
	log.Println("loaded config:" + configFile)
	vm.Run(configFileContents)
}

// getJSVM will return an Otto Object so you can run stuff, this will have loaded the definitions.
func getJSVM() *otto.Otto {

	vm := otto.New()
	runConfigInJSVM(vm)

	//	vm.Run(`
	//	    abc = 2 + 2;
	//	    console.log("The value of abc is " + abc); // 4
	//	`)
	return vm
}
func getRecordCSS(vm *otto.Otto, dbName string, row Site, fieldName string) string {
	rowAsJSON, err := json.Marshal(row)
	if err != nil {
		log.Println("can't jsonify row:" + err.Error())
		return ""
	}

	vm.Run(fmt.Sprintf("var res = getRowColor(%s,'%s','%s')", rowAsJSON, dbName, fieldName))
	valueAsString := "background-color:black"
	if value, err := vm.Get("res"); err == nil {
		if valueAsString, err := value.ToString(); err == nil {
			return valueAsString
		}
	}
	return valueAsString
}

// perhaps we can add:
// todo add flags  (in addition to env vars) for:
// port, host
// env vars are confusing, even for programmers, should we remove?
//
// todo: create a decent configuration file, which includes
// database connection for both tables (they may live in different hosts)
//
// *** comparison algorithm settings file (lua?? scripting)
//
// we need a gui to configure the program as well.
// maybe a config tab?
//
// our needs:
// provide sql where statements on both sides (config file)
//
// provide transformations (i.e. 100 becomes 10) lua-algo
//
// indicate which columns are important
//
// generate SQL for inserts, updates, maybe even DELETEs??!
//
// selecting which rows are important is also useful.
//
func main() {
	if len(os.Args) < 4 {
		log.Println("usage: db_diff db1 db2 configFile.js")
		return
	}
	db1name = os.Args[1]
	db2name = os.Args[2]
	configFile = os.Args[3]

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", fs) //http.StripPrefix("/public/", fs))
	http.Handle("/", fs)

	http.HandleFunc("/list", showDifferencesHandler)
	http.HandleFunc("/clear", clearChangesHandler)
	http.HandleFunc("/addChange", addChangeHandler)
	http.HandleFunc("/changes", listChangesHandler)

	fmt.Println("Server started, you probably want to go to http://localhost:3433/")
	err := http.ListenAndServe(":3433", nil) // 3433: diff in t9
	if err != nil {
		log.Fatal("Port is taken, maybe already running in another tab? out of sockets?: ", err)
	}
}

// http://stripesgenerator.com/
// https://astaxie.gitbooks.io/build-web-application-with-golang/en/03.2.html
