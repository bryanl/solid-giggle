package kep

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type Link struct {
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Links []Link

type KEP struct {
	Title             string     `json:"title,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	OwningSIG         string     `json:"owning-sig,omitempty"`
	ParticipatingSIGs []string   `json:"participating-sigs,omitempty"`
	Reviewers         []string   `json:"reviewers,omitempty"`
	Approvers         []string   `json:"approvers,omitempty"`
	Editor            string     `json:"editor,omitempty"`
	CreationDate      time.Time  `json:"creation-date,omitempty"`
	LastUpdated       *time.Time `json:"last-updated,omitempty"`
	Status            string     `json:"status,omitempty"`
	SeeAlso           Links      `json:"see-also,omitempty"`
	Content           string     `json:"-"`
}

func (k *KEP) String() (string, error) {
	var buf bytes.Buffer

	type Alias KEP

	creationDate := k.CreationDate.Format("2006-01-02")
	var lastUpdated string
	if k.LastUpdated != nil {
		lastUpdated = k.LastUpdated.Format("2006-01-02")
	}

	output := &struct {
		CreationDate string    `json:"creation-date,omitempty"`
		LastUpdated  string    `json:"last-updated,omitempty"`
		Date         time.Time `json:"date,omitempty"`
		Draft        bool      `json:"draft"`
		Tags         []string  `json:"tags,omitempty"`
		*Alias
	}{
		CreationDate: creationDate,
		LastUpdated:  lastUpdated,
		Date:         k.CreationDate,
		Draft:        false,
		Tags:         []string{k.OwningSIG},
		Alias:        (*Alias)(k),
	}

	buf.WriteString(fmt.Sprintf("%s\n", sectionMarker))

	data, err := yaml.Marshal(output)
	if err != nil {
		return "", err
	}

	fmt.Println(string(data))

	buf.Write(data)
	buf.WriteString(fmt.Sprintf("%s\n", sectionMarker))
	buf.WriteString(k.Content)

	return buf.String(), nil
}

func (k *KEP) MarshalJSON() ([]byte, error) {
	type Alias KEP

	creationDate := k.CreationDate.Format("2006-01-02")
	var lastUpdated string
	if k.LastUpdated != nil {
		lastUpdated = k.LastUpdated.Format("2006-01-02")
	}

	return json.Marshal(&struct {
		CreationDate string `json:"creation-date,omitempty"`
		LastUpdated  string `json:"last-updated,omitempty"`
		*Alias
	}{
		CreationDate: creationDate,
		LastUpdated:  lastUpdated,
		Alias:        (*Alias)(k),
	})
}

func (k *KEP) UnmarshalJSON(b []byte) error {
	type Alias KEP
	u := &struct {
		Authors      []interface{} `json:"authors"`
		Approvers    []interface{} `json:"approvers"`
		Reviewers    []interface{} `json:"reviewers"`
		Editor       *interface{}  `json:"editor"`
		RawSeeAlso   []string      `json:"see-also"`
		CreationDate string        `json:"creation-date"`
		LastUpdated  *string       `json:"last-updated"`
		*Alias
	}{
		Alias: (*Alias)(k),
	}
	if err := json.Unmarshal(b, &u); err != nil {
		return errors.Wrap(err, "unmarshal intermediary")
	}

	authors, err := extractUsers(u.Authors)
	if err != nil {
		return errors.Wrap(err, "extract author users")
	}
	k.Authors = authors

	approvers, err := extractUsers(u.Approvers)
	if err != nil {
		return errors.Wrap(err, "extract approver users")
	}
	k.Approvers = approvers

	reviewers, err := extractUsers(u.Reviewers)
	if err != nil {
		return errors.Wrap(err, "extract reviewer users")
	}
	k.Reviewers = reviewers

	if u.Editor != nil {
		editor, err := extractUser(*u.Editor)
		if err != nil {
			return errors.Wrap(err, "extract editor user")
		}
		k.Editor = editor
	}

	layout := "2006-01-02"
	creationDate, err := time.Parse(layout, u.CreationDate)
	if err != nil {
		return err
	}

	k.CreationDate = creationDate

	if u.LastUpdated != nil {
		lastUpdated, err := time.Parse(layout, *u.LastUpdated)
		if err != nil {
			return err
		}

		k.LastUpdated = &lastUpdated
	}

	for _, rawSeeAlso := range u.RawSeeAlso {
		matches := reLinkParts.FindAllStringSubmatch(rawSeeAlso, -1)
		for _, match := range matches {
			k.SeeAlso = append(k.SeeAlso, Link{Text: match[1], URL: match[2]})
		}
	}

	return nil
}

func extractUsers(rawUsers []interface{}) ([]string, error) {
	var users []string

	for _, rawUser := range rawUsers {
		user, err := extractUser(rawUser)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func extractUser(rawUser interface{}) (string, error) {
	switch t := rawUser.(type) {
	case string:
		return t, nil
	case map[string]interface{}:
		// user is an object. Get the approver name from the object
		name, ok := t["name"].(string)
		if !ok {
			return "", errors.Errorf("can't decipher approver: %v", t)
		}

		return name, nil
	default:
		return "", errors.Errorf("unsure what to do with: %v", t)
	}
}

const sectionMarker = "---"

// Read reads a kep from a reader.
func Read(r io.Reader) (*KEP, error) {
	sections, err := readSections(r)
	if err != nil {
		return nil, err
	}

	if len(sections) != 2 {
		return nil, errors.Errorf("invalid KEP (has %d sections)", len(sections))
	}

	kep, err := parseHeader(sections[0])
	if err != nil {
		return nil, err
	}

	kep.Content = sections[1]

	return kep, nil
}

func readSections(r io.Reader) ([]string, error) {
	currentLine := 0
	inHeader := false
	inContent := false

	var header strings.Builder
	var content strings.Builder

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		currentLine++
		text := scanner.Text()

		if text == sectionMarker {
			if !inHeader && !inContent {
				inHeader = true
				continue
			} else if inHeader && !inContent {
				inHeader = false
				inContent = true
				continue
			}
		}

		if inHeader {
			header.WriteString(fmt.Sprintf("%s\n", text))
		} else if inContent {
			content.WriteString(fmt.Sprintf("%s\n", text))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []string{
		strings.TrimSpace(header.String()),
		strings.TrimSpace(content.String()),
	}, nil
}

var reLinkParts = regexp.MustCompile(`^\[(.*?)\]\((.*?)\)$`)
var reArrayLink = regexp.MustCompile(`^(\s*-\s*)(\[.*?\))$`)

func parseHeader(in string) (*KEP, error) {
	in, err := fixHeaderQuotes(in)
	if err != nil {
		return nil, err
	}

	var kep KEP
	if err := yaml.Unmarshal([]byte(in), &kep); err != nil {
		return nil, err
	}

	return &kep, nil
}

func fixHeaderQuotes(in string) (string, error) {
	sb := strings.Builder{}

	scanner := bufio.NewScanner(strings.NewReader(in))
	for scanner.Scan() {
		text := scanner.Text()
		if reArrayLink.MatchString(text) {
			sb.WriteString(reArrayLink.ReplaceAllString(text, `$1"$2"`))
		} else {
			sb.WriteString(text)
		}
		sb.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return sb.String(), nil
}
