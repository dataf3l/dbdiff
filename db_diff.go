package main

import (
	"database/sql"
	"fmt"
	//"io"
	//"io/ioutil"
	"log"
	//"net/http"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"reflect"
	"strings"
	//"time"
)

func getDatabaseConnection(dbName string) *sql.DB {
	connectionString := getConnectionString(dbName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}
func checkErr(e error) {
	if e == nil {
		panic(e)
	}
}

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

// Site represents a record on the Site table, although it could probably be any table
type Site map[string]string

func getHeaders() string {
	dx := "<tr>"
	for _, h := range getFieldList() {
		dx += "<th>" + h + "</th>"
		dx += "<th>" + h + "</th>"
	}
	dx += "</tr>"
	return dx

}
func mkEmpty() Site {
	d := make(Site)
	for _, f := range getFieldList() {
		d[f] = "EMPTY"
	}
	return d
}
func getFieldList() []string {
	fields := []string{
		"id",
		//	"name",
		//	"site",
		//	"site_url",
		//	"company_name",
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
		"organization_logo",
		"script_template"}

	return fields
}
func compareRecords(site1 Site, site2 Site) (string, int) {
	fields := getFieldList()
	diffCount := 0
	var resultFields []string
	result := "<tr>"
	for _, fieldName := range fields {
		leftValue := site1[fieldName]
		rightValue := site2[fieldName]

		lv := "<xmp>" + leftValue + "</xmp>"
		rv := "<xmp>" + rightValue + "</xmp>"

		res := "<td class=ok>" + lv + "</td><td class=ok>" + rv + "</td>"

		if leftValue == "EMPTY" {
			res = "<td class=empty>&nbsp;</td><td>" + rv + "</td>"
			diffCount++
			resultFields = append(resultFields, res)
			continue
		}
		if rightValue == "EMPTY" {
			res = "<td>" + lv + "</td><td class=empty>&nbsp;</td>"
			diffCount++
			resultFields = append(resultFields, res)
			continue
		}

		if leftValue == rightValue {
			// Ignore differences
		} else {
			// Report Differences
			diffCount++
			res = "<td style='background-color:rgb(230,200,200)'>" + lv + "</td>"
			res += "<td style='background-color:rgb(230,200,200)' >" + rv + "</td>"
		}
		resultFields = append(resultFields, res)
	}
	result += strings.Join(resultFields, "")
	result += "</tr>\n"
	return result, diffCount
}
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
func getRecords(query string, dbName string) (map[string]Site, []string, error) {
	db := getDatabaseConnection(dbName)

	dbrows, err := db.Query(query)
	if err != nil {
		msg := fmt.Sprintf("fail : %s", err.Error())
		fmt.Println(msg)
		//io.WriteString(w, "Fail")
		db.Close()
		return make(map[string]Site), []string{}, err
	}
	result := make(map[string]Site)
	ids := []string{}
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
		ids = append(ids, record["id"])
		result[record["id"]] = record

		//io.WriteString(w, tmpl)
	}

	db.Close()
	return result, ids, nil
}
func getCombined(ids1 []string, ids2 []string) []string {
	// start:
	res := []string{}
	lid := 0
	rid := 0
	lmax := len(ids1)
	rmax := len(ids2)
	for {
		smallest := "NULL"
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

func main() {
	if len(os.Args) < 3 {
		log.Println("usage: db_diff db1 db2")
		return
	}
	db1name := os.Args[1]
	db2name := os.Args[2]
	q1 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site ORDER BY id"
	q2 := "SELECT " + strings.Join(getFieldList(), ",") + " FROM site ORDER BY id"

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

	html := "<table border=1 cellspacing=0 cellpadding=4>"
	html += getHeaders()
	total := 0
	for _, idx := range combined {
		leftRecord := mkEmpty()
		rightRecord := mkEmpty()

		if val, ok := leftRecordSet[idx]; ok {
			leftRecord = val
		}

		if val, ok := rightRecordSet[idx]; ok {
			rightRecord = val
		}

		r, c := compareRecords(leftRecord, rightRecord)
		total += c
		html += r
	}
	html += `</table>
	<style> 
	.empty {
		background-image: linear-gradient(45deg, #dbdbdb 25%, #f5e7e7 25%, #f5e7e7 50%, #dbdbdb 50%, #dbdbdb 75%, #f5e7e7 75%, #f5e7e7 100%);
background-size: 56.57px 56.57px;

	}

	td { 
		white-space:nowrap; 
		max-width:300px;
		overflow:hidden 
	} 
	.ok {
		background-color:rgb(200,230,200);
	}

	</style>`
	fmt.Println(html)

	fmt.Printf("Total: %d", total)
}

// http://stripesgenerator.com/
