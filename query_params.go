package didww

import (
	"fmt"
	"net/url"
	"strings"
)

// QueryParams builds query parameters for DIDWW API requests.
type QueryParams struct {
	values url.Values
}

// NewQueryParams creates a new QueryParams builder.
func NewQueryParams() *QueryParams {
	return &QueryParams{values: url.Values{}}
}

// Filter adds a filter parameter: filter[key]=value.
func (q *QueryParams) Filter(key, value string) *QueryParams {
	q.values.Set(fmt.Sprintf("filter[%s]", key), value)
	return q
}

// Sort sets the sort parameter with comma-separated fields.
// Prefix with "-" for descending order.
func (q *QueryParams) Sort(fields ...string) *QueryParams {
	q.values.Set("sort", strings.Join(fields, ","))
	return q
}

// Include sets the include parameter for related resources.
func (q *QueryParams) Include(relations ...string) *QueryParams {
	q.values.Set("include", strings.Join(relations, ","))
	return q
}

// Page sets pagination parameters.
func (q *QueryParams) Page(number, size int) *QueryParams {
	q.values.Set("page[number]", fmt.Sprintf("%d", number))
	q.values.Set("page[size]", fmt.Sprintf("%d", size))
	return q
}

// Fields sets sparse fieldsets for a resource type.
func (q *QueryParams) Fields(resourceType string, fields ...string) *QueryParams {
	q.values.Set(fmt.Sprintf("fields[%s]", resourceType), strings.Join(fields, ","))
	return q
}

// Encode returns the URL-encoded query string.
// Brackets are kept unescaped for JSON:API compatibility (filter[key]=value).
func (q *QueryParams) Encode() string {
	encoded := q.values.Encode()
	encoded = strings.ReplaceAll(encoded, "%5B", "[")
	encoded = strings.ReplaceAll(encoded, "%5D", "]")
	return encoded
}
