package newrelic

import (
	"strings"
)

// Link represents a single link from a Link header
type Link struct {
	URL string
	Rel string
}

// Links represents a collection of links parsed from a Link header
type Links []Link

// ParseLinkHeader parses an HTTP Link header into a collection of Links.
// Link headers follow RFC 5988/8288 format:
// Link: <url1>; rel="relation1", <url2>; rel="relation2"
func ParseLinkHeader(header string) Links {
	var links Links

	if header == "" {
		return links
	}

	// Split by comma to get individual links
	parts := strings.Split(header, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		link := Link{}

		// Split by semicolon to separate URL from parameters
		segments := strings.Split(part, ";")
		if len(segments) == 0 {
			continue
		}

		// Extract URL (remove < and >)
		urlPart := strings.TrimSpace(segments[0])
		if strings.HasPrefix(urlPart, "<") && strings.HasSuffix(urlPart, ">") {
			link.URL = urlPart[1 : len(urlPart)-1]
		} else {
			continue
		}

		// Parse parameters (looking for rel="...")
		for i := 1; i < len(segments); i++ {
			param := strings.TrimSpace(segments[i])
			if strings.HasPrefix(param, "rel=") {
				// Extract rel value (remove quotes if present)
				relValue := strings.TrimPrefix(param, "rel=")
				relValue = strings.Trim(relValue, "\"'")
				link.Rel = relValue
				break
			}
		}

		if link.URL != "" && link.Rel != "" {
			links = append(links, link)
		}
	}

	return links
}

// FilterByRel returns all links that match the specified relation type
func (links Links) FilterByRel(rel string) Links {
	var filtered Links
	for _, link := range links {
		if link.Rel == rel {
			filtered = append(filtered, link)
		}
	}
	return filtered
}
