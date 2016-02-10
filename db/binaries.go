package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash/native"
	"github.com/Sirupsen/logrus"
	"github.com/hobeone/gonab/types"
)

func (d *Handle) getRegexesForGroups(groupNames []string, includeWildCard bool) ([]types.Regex, error) {
	var regexesToUse []types.Regex

	if includeWildCard {
		groupNames = append(groupNames, ".*")
	}
	err := d.DB.Where("group_name in (?)", groupNames).Order("ordinal").Find(&regexesToUse).Error
	if err != nil {
		return nil, fmt.Errorf("Error getting regexes to use: %v", err)
	}

	var goodRegexes []types.Regex
	for _, r := range regexesToUse {
		err = r.Compile()
		if err != nil {
			logrus.Errorf("Regex %d compile error: %v", r.ID, err)
			continue
		}
		goodRegexes = append(goodRegexes, r)
	}

	return goodRegexes, nil
}

// MakeBinaries scans for new parts (where binary_id is NULL) and tries to
// assemble them into Binaries.
//
// It does this by applying Regex patterns to their subjects and extracting
// information.
func (d *Handle) MakeBinaries() error {
	var partGroups []string
	err := d.DB.Model(&types.Part{}).Group("group_name").Pluck("group_name", &partGroups).Error
	if err != nil {
		return fmt.Errorf("Error getting group names for new parts: %v", err)
	}

	regexesToUse, err := d.getRegexesForGroups(partGroups, true)
	if len(regexesToUse) < 1 {
		return fmt.Errorf("No regexes found, have you imported any?")
	}

	var parts []types.Part
	var partCount int64
	err = d.DB.Where("group_name in (?) AND binary_id is NULL", partGroups).Find(&parts).Count(&partCount).Error
	if err != nil {
		return err
	}
	logrus.Infof("Found %d parts to process.", partCount)

	binaries := map[string]*types.Binary{}

	for _, p := range parts {
		matched := false
		for _, r := range regexesToUse {
			if !strings.HasPrefix(p.GroupName, r.GroupName) && r.GroupName != ".*" {
				continue
			}
			matches, err := matchPart(r.Compiled, &p)
			if err != nil {
				continue
			}
			logrus.Debugf("Found match with regex %d (%s): %v", r.ID, r.GroupName, r.Regex)
			matched = true
			partcounts := strings.SplitN(matches["parts"], "/", 2)

			binhash := makeBinaryHash(matches["name"], p.GroupName, p.From, partcounts[1])
			if bin, ok := binaries[binhash]; ok {
				bin.Parts = append(bin.Parts, p)
			} else {
				b, err := d.FindBinaryByHash(binhash)
				if err != nil {
					logrus.Debugf("New binary found: %s", matches["name"])
					totalparts, _ := strconv.Atoi(partcounts[1])
					binaries[binhash] = &types.Binary{
						Hash:       binhash,
						Name:       matches["name"],
						Posted:     p.Posted,
						From:       p.From,
						Parts:      []types.Part{p},
						GroupName:  p.GroupName,
						TotalParts: totalparts,
					}
				} else {
					b.Parts = append(b.Parts, p)
				}
			}

			err = d.DB.Save(binaries[binhash]).Error
			if err != nil {
				return err
			}

			break
		}
		if !matched {
			logrus.Infof("Couldn't match %s with any regex, deleting.", p.Subject)
			d.DB.Delete(p)
		}
	}
	return nil
}

func makeBinaryHash(name, group, from, totalParts string) string {
	h := xxhash.New64()
	h.Write([]byte(name + group + from + totalParts))
	return fmt.Sprintf("%x", h.Sum64())
}

// partRegex is the fallback regex to find parts.
var partRegex = regexp.MustCompile(`(?i)[\[\( ]((\d{1,3}\/\d{1,3})|(\d{1,3} of \d{1,3})|(\d{1,3}-\d{1,3})|(\d{1,3}~\d{1,3}))[\)\] ]`)

func matchPart(r *types.RegexpUtil, p *types.Part) (map[string]string, error) {
	m := r.FindStringSubmatchMap(p.Subject)
	for k, v := range m {
		m[k] = strings.TrimSpace(v)
	}
	// fill name if reqid is available
	if reqid, ok := m["reqid"]; ok {
		if _, okname := m["name"]; !okname {
			m["name"] = reqid
		}
	}

	// Generate a name if we don't have one
	if _, ok := m["name"]; !ok {
		matchvalues := make([]string, len(m))
		i := 0
		for _, v := range m {
			matchvalues[i] = v
			i++
		}
		m["name"] = strings.Join(matchvalues, " ")
	}

	// Look for parts manually if the regex didn't return some
	if _, ok := m["parts"]; !ok {
		partmatch := partRegex.FindStringSubmatch(p.Subject)
		if partmatch != nil {
			m["parts"] = partmatch[1]
		}
	}
	if !hasNameAndParts(m) {
		return m, fmt.Errorf("Couldn't find Name and Parts for %s\n", p.Subject)
	}

	// Clean name of '-', '~', ' of '
	if strings.Index(m["parts"], "/") == -1 {
		m["parts"] = strings.Replace(m["parts"], "-", "/", -1)
		m["parts"] = strings.Replace(m["parts"], "~", "/", -1)
		m["parts"] = strings.Replace(m["parts"], " of ", "/", -1)
		m["parts"] = strings.Replace(m["parts"], "[", "", -1)
		m["parts"] = strings.Replace(m["parts"], "]", "", -1)
		m["parts"] = strings.Replace(m["parts"], "(", "", -1)
		m["parts"] = strings.Replace(m["parts"], ")", "", -1)
	}

	if strings.Index(m["parts"], "/") == -1 {
		return nil, fmt.Errorf("Couldn't find valid parts information for %s (%s didn't include /)\n", p.Subject, m["parts"])
	}
	return m, nil
}

func hasNameAndParts(m map[string]string) bool {
	var nameok, partok bool
	if _, nameok = m["name"]; nameok {
		nameok = m["name"] != ""
	}
	if _, partok = m["parts"]; partok {
		partok = m["parts"] != ""
	}
	return nameok && partok
}
