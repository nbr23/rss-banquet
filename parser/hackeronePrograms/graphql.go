package hackeronePrograms

import (
	"encoding/json"
	"net/http"
	"strings"
)

func programsFeedQuery() (*http.Response, error) {
	query := `query DiscoveryQuery($query: OpportunitiesQuery!, $filter: QueryInput!, $from: Int, $size: Int, $sort: [SortInput!], $post_filters: OpportunitiesFilterInput) {
        me {
            id
            ...OpportunityListMe
            __typename
        }
        opportunities_search(query: $query
        filter: $filter
        from: $from
        size: $size
        sort: $sort
        post_filters: $post_filters) {
            nodes {
                ... on OpportunityDocument {
                    id
                    handle
                    __typename
                }
                ...OpportunityList
                __typename
            }
            total_count
            __typename
        }
    }
    
    fragment OpportunityListMe on User {
        id
        ...OpportunityCardMe
        __typename
    }
    
    fragment OpportunityCardMe on User {
        id
        ...BookmarkMe
        __typename
    }
    
    fragment BookmarkMe on User {
        id
        __typename
    }
    
    fragment OpportunityList on OpportunityDocument {
        id
        ...OpportunityCard
        __typename
    }
    
    fragment OpportunityCard on OpportunityDocument {
        id
        team_id
        name
        handle
        profile_picture
        triage_active
        publicly_visible_retesting
        allows_private_disclosure
        allows_bounty_splitting
        launched_at
        state
        offers_bounties
        last_updated_at
        currency
        team_type
        minimum_bounty_table_value
        maximum_bounty_table_value
        cached_response_efficiency_percentage
        first_response_time
        structured_scope_stats
        show_response_efficiency_indicator
        submission_state
        resolved_report_count
        campaign {
            id
            campaign_type
            start_date
            end_date
            critical
            target_audience
            __typename
        }
        gold_standard
        awarded_report_count
        awarded_reporter_count
        h1_clear
        idv
        __typename
    }`

	variables := map[string]any{
		"query": map[string]any{},
		"sort": []map[string]any{
			map[string]any{
				"field":     "launched_at",
				"direction": "DESC",
			},
		},
		"product_area": "opportunity_discovery",
		"filter": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{
						"bool": map[string]any{
							"must_not": map[string]any{
								"term": map[string]string{
									"team_type": "Engagements::Assessment",
								},
							},
						},
					},
				},
			},
		},
		"product_feature": "search",
		"size":            24,
		"from":            0,
		"post_filters": map[string]bool{
			"bookmarked":     false,
			"campaign_teams": false,
		},
	}

	gql := map[string]any{
		"query":     query,
		"variables": variables,
	}

	client := &http.Client{}
	jsonValue, _ := json.Marshal(gql)

	req, err := http.NewRequest(
		"POST",
		"https://hackerone.com/graphql",
		strings.NewReader(string(jsonValue)),
	)

	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}
