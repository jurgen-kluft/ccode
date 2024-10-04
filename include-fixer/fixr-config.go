package fixr

import (
	"path/filepath"
	"regexp"
	"strings"
)

type IncludeGuardConfig struct {
	UseFilename               bool
	RemovePrefix              string
	RemoveSuffix              string
	AddPrefix                 string
	AddSuffix                 string
	IncludeGuardIfNotDefRegex *regexp.Regexp
	IncludeGuardDefineRegex   *regexp.Regexp
}

func NewIncludeGuardConfig() *IncludeGuardConfig {
	d := &IncludeGuardConfig{}
	d.UseFilename = true
	d.RemovePrefix = ""
	d.RemoveSuffix = ""
	d.AddPrefix = "__CRED_"
	d.AddSuffix = "__"
	d.IncludeGuardIfNotDefRegex, _ = regexp.Compile(`#ifndef\s+(.*)\s*`)
	d.IncludeGuardDefineRegex, _ = regexp.Compile(`#define\s+(.*)\s*`)
	return d
}

func (d *IncludeGuardConfig) ensureDefineIsValidString(define string) string {
	valid := strings.Map(func(r rune) rune {
		if 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, define)
	return valid
}

func (d *IncludeGuardConfig) modifyDefine(guard string, _filepathOfFileToFix string) string {
	if d.UseFilename {
		define := strings.ToUpper(filepath.Base(_filepathOfFileToFix))
		define = d.ensureDefineIsValidString(define)
		return d.AddPrefix + define + d.AddSuffix // Add the prefix and suffix
	} else {
		for strings.HasPrefix(guard, d.RemovePrefix) {
			guard = guard[len(d.RemovePrefix):]
		}
		for strings.HasSuffix(guard, d.RemoveSuffix) {
			guard = guard[:len(guard)-len(d.RemoveSuffix)]
		}
		return d.AddPrefix + guard + d.AddSuffix // Add the prefix and suffix
	}
}

type IncludeDirectiveConfig struct {
	IncludeDirectiveRegex *regexp.Regexp
}

func NewIncludeDirectiveConfig() *IncludeDirectiveConfig {
	d := &IncludeDirectiveConfig{}
	d.IncludeDirectiveRegex, _ = regexp.Compile(`#include\s+(["<])(.*)([">])`)
	return d
}
