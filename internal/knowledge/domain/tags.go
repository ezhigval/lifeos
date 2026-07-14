package domain

import (
	"regexp"
	"strings"
)

var hashtagRe = regexp.MustCompile(`(?i)#([\p{L}\p{N}_-]+)`)

func NormalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "#")))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func ExtractHashtags(body string) (cleanBody string, tags []string) {
	matches := hashtagRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return strings.TrimSpace(body), nil
	}
	tagSet := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		tagSet[strings.ToLower(m[1])] = struct{}{}
	}
	clean := hashtagRe.ReplaceAllString(body, " ")
	clean = strings.Join(strings.Fields(clean), " ")
	tags = make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return clean, NormalizeTags(tags)
}
