// The MIT License (MIT)
//
// Copyright (c) 2015 Marc Rohlfs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
    "encoding/csv"
    "encoding/xml"
    "io/ioutil"
    "net/http"
    "github.com/everdev/mack"
    "github.com/jimlawless/cfg"
)

const TMP_TYME_EXPORT_CSV_FILE_NAME = "tmp-tyme-export.csv"

type TimeEntry struct {
    XMLName   xml.Name `xml:"time-entry"`
    Date      string   `xml:"date-at"`
    Minutes   int      `xml:"minutes"`
    Note      string   `xml:"note"`
    ServiceId int      `xml:"service-id"`
    ProjectId int      `xml:"project-id"`
}

type Projects struct {
    XMLName xml.Name  `xml:"projects"`
    Project []Project `xml:"project"`
}

type Project struct {
    XMLName      xml.Name `xml:"project"`
    Id           int      `xml:"id"`
    Name         string   `xml:"name"`
    CustomerId   int      `xml:"customer-id"`
    CustomerName string   `xml:"customer-name"`
}

type Services struct {
    XMLName xml.Name  `xml:"services"`
    Service []Service `xml:"service"`
}

type Service struct {
    Id   int    `xml:"id"`
    Name string `xml:"name"`
}

func main() {

    var startDate string
    if len(os.Args) >= 2 {
        startDate = os.Args[1]
    } else {
        startDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
    }

    var endDate string
    if len(os.Args) >= 3 {
        endDate = os.Args[2]
    } else {
        endDate = time.Now().Format("2006-01-02")
    }

    config, err := cfg.LoadNewMap(os.Getenv("HOME") + "/tyme2mite.cfg")
    checkErr(err)
    miteImportActive := strings.ToLower(config["mite_import_active"]) == "true"
    miteBaseUrl := string(config["mite_base_url"])
    miteApiKey := string(config["mite_api_key"])
    miteRequestUrlPattern := miteBaseUrl + "/[RESOURCE_ID]?api_key=" + miteApiKey
    miteTimeEntriesUrl := strings.Replace(miteRequestUrlPattern, "[RESOURCE_ID]", "time_entries.xml", 1)
    miteProjectsUrl := strings.Replace(miteRequestUrlPattern, "[RESOURCE_ID]", "projects.xml", 1)
    miteServicesUrl := strings.Replace(miteRequestUrlPattern, "[RESOURCE_ID]", "services.xml", 1)

    if miteImportActive == false {
        fmt.Print("[DRY RUN] ")
    }
    fmt.Println("Transferring time entries from " + startDate + " to " + endDate + " ...")

    projects := Projects{}
    res, err := http.Get(miteProjectsUrl)
    checkErr(err)
    xmlBody, err := ioutil.ReadAll(res.Body)
    checkErr(err)
    xml.Unmarshal(xmlBody, &projects)
    res.Body.Close()

    services := Services{}
    resSrv, err := http.Get(miteServicesUrl)
    checkErr(err)
    xmlBodySrv, err := ioutil.ReadAll(resSrv.Body)
    checkErr(err)
    xml.Unmarshal(xmlBodySrv, &services)
    resSrv.Body.Close()

    err = mack.Tell("Tyme",
        "set ex to make new export",
        "set startDate of ex to (date \"" + startDate + "\")",
        "set endDate of ex to (date \"" + endDate + "\")",
        "set exportFormat of ex to csv",
        "set exportFileName of ex to \"" + TMP_TYME_EXPORT_CSV_FILE_NAME + "\"",
        "save export ex")
    checkErr(err)

    tmpTymeExportCsvFilePath := os.Getenv("HOME") + "/Downloads/" + TMP_TYME_EXPORT_CSV_FILE_NAME
    csvfile, err := os.Open(tmpTymeExportCsvFilePath)
    checkErr(err)

    defer csvfile.Close()
    os.Remove(tmpTymeExportCsvFilePath)

    csvReader := csv.NewReader(csvfile)
    checkErr(err)

    csvReader.Comma = ';'
    csvReader.FieldsPerRecord = -1

    csvColHeaders, err := csvReader.Read()
    checkErr(err)

    assert("Date", csvColHeaders[0])
    assert("Project", csvColHeaders[1])
    assert("Task", csvColHeaders[2])
    assert("Duration", csvColHeaders[6])
    assert("Notes", csvColHeaders[9])

    rawCSVdata, err := csvReader.ReadAll()
    checkErr(err)

    var timeEntries []TimeEntry
    for _, each := range rawCSVdata {

        date := each[0]

        duration := strings.Split(each[6], ":")
        hours, err := strconv.Atoi(duration[0])
        checkErr(err)
        minutes, err := strconv.Atoi(duration[1])
        checkErr(err)
        minutes = hours * 60 + minutes

        var projectId int
        customerProject := strings.Split(each[1], "|")
        customer := strings.TrimSpace(customerProject[0])
        project := strings.TrimSpace(customerProject[1])
        for idx := 0; idx < len(projects.Project); idx++ {
            if customer == projects.Project[idx].CustomerName && project == projects.Project[idx].Name {
                projectId = projects.Project[idx].Id
                break
            }
        }

        var notePrefix string
        var noteText string
        var service string
        taskService := strings.Split(each[2], "|")
        if len(taskService) > 1 {
            notePrefix = strings.TrimSpace(taskService[0]) + ": "
            noteText = each[9]
            service = strings.TrimSpace(taskService[1])
        } else {
            notePrefix = ""
            noteText = each[9]
            service = strings.TrimSpace(taskService[0])
        }

        var serviceId int
        for idx := 0; idx < len(services.Service); idx++ {
            if service == services.Service[idx].Name {
                serviceId = services.Service[idx].Id
                break
            }
        }

        cumulateTimeEntryIndex := -1
        for idx := 0; idx < len(timeEntries); idx++ {
            if timeEntries[idx].Date == date && timeEntries[idx].ProjectId == projectId && timeEntries[idx].ServiceId == serviceId {
                if len(notePrefix) == 0 || strings.HasPrefix(timeEntries[idx].Note, notePrefix) {
                    cumulateTimeEntryIndex = idx
                    break;
                }
            }
        }

        if cumulateTimeEntryIndex == -1 {
            var timeEntry TimeEntry
            timeEntry.Date = date
            timeEntry.Minutes = minutes
            timeEntry.Note = notePrefix + noteText
            timeEntry.ProjectId = projectId
            timeEntry.ServiceId = serviceId
            timeEntries = append(timeEntries, timeEntry)
        } else {
            timeEntries[cumulateTimeEntryIndex].Minutes += minutes
            timeEntries[cumulateTimeEntryIndex].Note = timeEntries[cumulateTimeEntryIndex].Note + ", " + noteText
        }
    }

    for idx := 0; idx < len(timeEntries); idx++ {

        xmlBody, err := xml.MarshalIndent(timeEntries[idx], "", "  ")
        checkErr(err)

        var xmlString = string(xmlBody)
        fmt.Println(xmlString)

        if miteImportActive {
            res, err := http.Post(miteTimeEntriesUrl, "application/xml", strings.NewReader(string(xmlBody)))
            checkErr(err)
            fmt.Print("Import result: ")
            fmt.Println(res)
        }

        fmt.Println()
    }

    if miteImportActive == false {
        fmt.Print("[DRY RUN] ")
    }
    fmt.Println("Transferred time entries from " + startDate + " to " + endDate)
}

func assert(expected string, actual string) {
    if expected != actual {
        panic("Unexpected column: " + actual)
    }
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
