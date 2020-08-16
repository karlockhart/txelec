package txelec

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
)

// DownloadURL holds a URL to a zipped CSV file and metadata to allow sorting
type DownloadURL struct {
	URL  string
	Name string
	Type string
	Day  int
	Time int
}

// GetLatestDLURL returns the most recent download URL index at url
func GetLatestDLURL(url string) (DownloadURL, error) {
	var d *DownloadURL
	dls, err := GetDLURLs(url)
	if err != nil {
		return *d, err
	}

	for n := range dls {
		dl := dls[n]

		if d != nil {
			logrus.Debug(dl.Day, " v ", d.Day, "       ", dl.Time, " v ", d.Time)
		}

		if d == nil || (dl.Day >= d.Day && dl.Time >= d.Time) {
			d = &dl
			logrus.Debug("set latest to ", d.Name)
		}
	}
	return *d, err
}

// GetDLURLs all download URLs from the index at url
func GetDLURLs(url string) ([]DownloadURL, error) {
	var dls []DownloadURL
	g, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	g.Find("tr").Each(func(i int, s *goquery.Selection) {
		dl := DownloadURL{}
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			u, e := s.Find("a").Attr("href")
			if e {
				if len(u) > 1 || string(u[0]) == "/" {
					dl.URL = fmt.Sprintf("%s%s", viper.GetString("download.dl_root"), string(u))
				} else {
					dl.URL = string(u)
				}
			}

			if s.HasClass("labelOptional_ind") {
				dl.Name = s.Text()
			}
		})
		parts := strings.Split(dl.Name, ".")
		if len(parts) > 2 {
			fnParts := strings.Split(parts[len(parts)-2], "_")
			if fnParts[len(fnParts)-1] == "csv" {
				dl.Type = "csv"
				d, err := strconv.Atoi(fnParts[len(fnParts)-3])
				if err != nil {
					return
				}
				t, err := strconv.Atoi(fnParts[len(fnParts)-2])
				if err != nil {
					return
				}
				dl.Day = d
				dl.Time = t
				dls = append(dls, dl)
			}
		}
	})
	return dls, nil
}

// GetDataFromZippedCSV returns a slice of string mapped strings from a zipped CSV
func GetDataFromZippedCSV(url string) ([]map[string]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	zbody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	zr := bytes.NewReader(zbody)

	r, err := zip.NewReader(zr, int64(len(zbody)))
	if err != nil {
		logrus.Fatal(err)
	}

	var recmap []map[string]string

	for _, f := range r.File {
		logrus.Debug(f.Name)
		fh, err := f.Open()
		if err != nil {
			return nil, err
		}
		csvr := csv.NewReader(fh)
		recs, err := csvr.ReadAll()
		if err != nil {
			return nil, err
		}

		headers := recs[0]

		for _, row := range recs[1:] {
			rm := make(map[string]string)
			for k, v := range row {
				rm[headers[k]] = strings.Trim(v, " \t\n")
			}
			recmap = append(recmap, rm)
		}

	}
	return recmap, err
}
