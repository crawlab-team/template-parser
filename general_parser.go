package parser

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type GeneralParser struct {
	tagPattern   string
	tagRegexp    *regexp.Regexp
	template     string
	indexes      [][]int
	matches      [][]string
	placeholders []string
	variables    []Variable
}

func (p *GeneralParser) Parse(template string) (err error) {
	p.template = template
	p.indexes = p.tagRegexp.FindAllStringIndex(template, -1)
	p.matches = p.tagRegexp.FindAllStringSubmatch(template, -1)
	for _, arr := range p.matches {
		p.placeholders = append(p.placeholders, arr[1])
	}

	return nil
}

func (p *GeneralParser) Render(args ...interface{}) (content string, err error) {
	// validate
	if len(args) == 0 {
		return "", errors.New("no arguments")
	}

	// first argument
	arg := args[0]

	// content
	content = p.template

	// old strings
	var oldStrList []string
	for _, index := range p.indexes {
		// old string
		oldStr := content[index[0]:index[1]]
		oldStrList = append(oldStrList, oldStr)
	}

	// iterate placeholders
	for i, placeholder := range p.placeholders {
		// variable
		v, err := NewVariable(arg, placeholder)
		if err != nil {
			return "", err
		}

		// value
		value, err := v.GetValue()
		if err != nil {
			return "", err
		}

		// old string
		oldStr := oldStrList[i]

		// new string
		var newStr string
		switch value.(type) {
		case string:
			newStr = value.(string)
		default:
			newStrBytes, err := json.Marshal(value)
			if err != nil {
				return "", err
			}
			newStr = string(newStrBytes)
		}

		// replace old string with new string
		content = strings.Replace(content, oldStr, newStr, 1)
	}

	return content, nil
}

func (p *GeneralParser) GetPlaceholders() (placeholders []string) {
	return p.placeholders
}

func NewGeneralParser() (p Parser, err error) {
	tagPattern := "\\{\\{ *([\\$\\.\\w_]+) *\\}\\}"
	tagRegexp, err := regexp.Compile(tagPattern)
	if err != nil {
		return nil, err
	}
	p = &GeneralParser{
		tagPattern: tagPattern,
		tagRegexp:  tagRegexp,
	}

	return p, nil
}
