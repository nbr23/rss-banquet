package nytimes

type GraphQLResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	AnyWork AnyWork `json:"anyWork"`
}

type AnyWork struct {
	ContentSearch ContentSearch `json:"contentSearch"`
	DisplayName   string        `json:"displayName,omitempty"`
}

type ContentSearch struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Edges []Edge `json:"edges"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Headline       Headline `json:"headline"`
	Summary        string   `json:"summary,omitempty"`
	URL            string   `json:"url,omitempty"`
	FirstPublished string   `json:"firstPublished,omitempty"`
}

type Headline struct {
	Typename string `json:"__typename,omitempty"`
	Default  string `json:"default"`
}
