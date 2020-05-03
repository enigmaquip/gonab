package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/enigmaquip/gonab/db"
	"github.com/enigmaquip/gonab/types"
	"github.com/hobeone/rss2go/httpclient"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var nnpplusRegexURL = `http://www.newznab.com/getregex.php?newznabID=%s`

//RegexImporter comment
type RegexImporter struct{}

var regexWithOption = regexp.MustCompile(`\/(\w+)$`)

func cleanRegex(re string) string {
	regex := strings.TrimSpace(re)
	regex = strings.Replace(regex, `\\`, `\`, -1)
	regex = strings.TrimLeft(regex, `/`)
	regexOpts := regexWithOption.FindStringSubmatch(regex)
	if regexOpts != nil {
		if strings.Contains(regexOpts[1], "i") {
			regex = fmt.Sprintf("(?i)") + regex // add re2 ignore case option
		}
		regex = strings.TrimRight(regex, regexOpts[0])
	} else {
		regex = strings.TrimRight(regex, "/")
	}
	return regex
}

// parsed order:
// 0 full string
// 1 id
// 2 group
// 3 regex
// 4 poster
// 5 ordinal
// 6 status
// 7 description
// 8 categoryID
func newzNabRegexToRegex(parsed []string) (*types.Regex, error) {
	if len(parsed) != 9 {
		return nil, fmt.Errorf("parsed newznab regex should be 8 items")
	}
	status, err := strconv.ParseBool(parsed[6])
	if err != nil {
		logrus.Errorf("Couldn't parse %s to bool, assuming true", parsed[5])
		status = true
	}
	id, err := strconv.Atoi(parsed[1])
	if err != nil {
		logrus.Errorf("Couldn't parse id %s, skipping", parsed[1])
	}
	ord, err := strconv.Atoi(parsed[5])
	if err != nil {
		logrus.Errorf("Couldn't parse ordinal %s, skipping", parsed[4])
	}
	regex := cleanRegex(parsed[3])
	dbregex := types.Regex{
		ID:          id,
		GroupRegex:  parsed[2],
		Regex:       regex,
		Status:      status,
		Description: parsed[7],
		Ordinal:     ord,
	}
	_, err = regexp.Compile(dbregex.Regex)
	if err != nil {
		return nil, fmt.Errorf("Error compiling regex, skipping: %v", err)
	}
	return &dbregex, nil
}

var splitRegex = regexp.MustCompile(`\((\d+), \'(.*)\', \'(.*)\', (.*), (\d+), (\d+), (.*), (.*)\);$`)

func parseNewzNabRegexes(b []byte) ([]*types.Regex, error) {
	r := bufio.NewReader(bytes.NewReader(b))
	newregexes := []*types.Regex{}
	lines := 0
	goodregex := 0
	badregex := 0
	for {
		lines++
		record, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if lines <= 2 {
			// Skip the revision and delete lines
			continue
		}
		record = strings.TrimSpace(record)
		matches := splitRegex.FindStringSubmatch(record)

		if len(matches) != 9 {
			logrus.Errorf("Invalid line in regex file: %s", record)
			badregex++
			continue
		} else {
			newregex, err := newzNabRegexToRegex(matches)
			if err != nil {
				logrus.Errorf("Couldn't create Regex from %v: %v", record, err)
				badregex++
				continue
			}
			newregexes = append(newregexes, newregex)
			goodregex++
		}
	}
	logrus.Infof("Found %d regexs, couldn't parse %d of them.", goodregex+badregex, badregex)
	return newregexes, nil
}

// Field Order
// 0 id
// 1 group_regex
// 2 regex
// 3 status
// 4 description
// 5 ordinal
func nzedbRegexToRegex(parsed []string, kind string) (*types.Regex, error) {
	if len(parsed) != 6 {
		return nil, fmt.Errorf("nzedb regexes must be 6 items")
	}
	status, err := strconv.ParseBool(parsed[3])
	if err != nil {
		logrus.Errorf("Couldn't parse %s to bool, assuming true", parsed[3])
		status = true
	}
	id, err := strconv.Atoi(parsed[0])
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse id %s, skipping", parsed[0])
	}
	ord, err := strconv.Atoi(parsed[5])
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse ordinal %s, skipping", parsed[5])
	}
	regex := cleanRegex(parsed[2])
	groupname := strings.Replace(parsed[1], `\\`, "", -1)
	dbregex := types.Regex{
		ID:          id,
		GroupRegex:  groupname,
		Regex:       regex,
		Status:      status,
		Description: parsed[4],
		Ordinal:     ord,
		Kind:        kind,
	}
	_, err = regexp.Compile(dbregex.Regex)
	if err != nil {
		return nil, fmt.Errorf("Error compiling regex, skipping: %v", err)
	}
	_, err = regexp.Compile(dbregex.GroupRegex)
	if err != nil {
		return nil, fmt.Errorf("Error compiling group regex, skipping: %v", err)
	}

	return &dbregex, nil
}

// Format is tab separated:
// id, group_regex, regex, status, description, ordinal
func parseNzedbRegexes(b []byte, kind string) ([]*types.Regex, error) {
	r := bufio.NewReader(bytes.NewReader(b))
	newregexes := []*types.Regex{}
	lines := 0

	for {
		lines++
		record, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if lines <= 1 {
			continue
		}
		record = strings.TrimSpace(record)
		fields := strings.Split(record, "\t")
		regex, err := nzedbRegexToRegex(fields, kind)
		if err != nil {
			logrus.Errorf("Error parsing nZEDb regex (line %d): %v", lines, err)
			continue
		}
		newregexes = append(newregexes, regex)
	}
	return newregexes, nil
}

func getURL(url string) ([]byte, error) {
	// Defaults to 1 second for connect and read
	connectTimeout := (5 * time.Second)
	readWriteTimeout := (15 * time.Second)

	client := httpclient.NewTimeoutClient(connectTimeout, readWriteTimeout)

	resp, err := client.Get(url)

	if err != nil {
		logrus.Infof("Error getting %s: %s", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("feed %s returned a non 200 status code: %s", url, resp.Status)
		logrus.Error(err)
		return nil, err
	}
	var b []byte
	if resp.ContentLength > 0 {
		b = make([]byte, resp.ContentLength)
		_, err := io.ReadFull(resp.Body, b)
		if err != nil {
			return nil, fmt.Errorf("error reading response for %s: %s", url, err)
		}
	} else {
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response for %s: %s", url, err)
		}
	}
	return b, nil
}

func (regeximporter *RegexImporter) run(c *kingpin.ParseContext) error {
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Infof("Reading config %s\n", *configfile)

	cfg := loadConfig(*configfile)
	url := cfg.Regex.URL
	logrus.Infof("Crawling %v", url)

	b, err := getURL(url)
	if err != nil {
		return err
	}
	var collectionRegexRaw []byte
	if cfg.Regex.CollectionURL != "" {
		logrus.Infof("CollectionURL set, crawling: %s", cfg.Regex.CollectionURL)
		collectionRegexRaw, err = getURL(cfg.Regex.CollectionURL)
		if err != nil {
			return err
		}
	}

	var regexes []*types.Regex
	var collectionRegexes []*types.Regex
	switch cfg.Regex.Type {
	case "nnplus":
		regexes, err = parseNewzNabRegexes(b)
	case "nzedb":
		regexes, err = parseNzedbRegexes(b, "release")
		collectionRegexes, err = parseNzedbRegexes(collectionRegexRaw, "collection")
	default:
		return fmt.Errorf("Unknown Regex type: %s", cfg.Regex.Type)
	}
	if err != nil {
		return err
	}
	logrus.Infof("Parsed %d regexes from %s", len(regexes), url)

	dbh := db.NewDBHandle(cfg.DB.Name, cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Verbose)
	newcount := 0
	tx := dbh.DB.Begin()
	tx.Where("id < ?", 100000).Delete(&types.Regex{})
	for _, dbregex := range regexes {
		err = tx.Create(dbregex).Error
		if err != nil {
			logrus.Errorf("Error saving regex: %v", err)
			tx.Rollback()
			return err
		}
		newcount++
	}
	tx.Where("id < ?", 100000).Delete(&types.Regex{Kind: "collection"})
	for _, dbregex := range collectionRegexes {
		err = tx.Create(dbregex).Error
		if err != nil {
			logrus.Errorf("Error saving collection regex: %v", err)
			tx.Rollback()
			return err
		}
		newcount++
	}
	tx.Commit()
	logrus.Infof("Saved %d regexes.", newcount)
	return nil
}
