package hackerone

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nbr23/atomic-banquet/parser"
)

func hacktivityFeedQuery(options *parser.Options) (*http.Response, error) {
	disclosed_only := options.Get("disclosed_only").(bool)
	reports_count := options.Get("reports_count").(int)

	query := `query HacktivityPageQuery($querystring: String, $orderBy: HacktivityItemOrderInput, $secureOrderBy: FiltersHacktivityItemFilterOrder, $where: FiltersHacktivityItemFilterInput, $count: Int, $cursor: String) {
        me {
            id
            __typename
        }
        hacktivity_items(first: $count
        after: $cursor
        query: $querystring
        order_by: $orderBy
        secure_order_by: $secureOrderBy
        where: $where) {
            ...HacktivityList
            __typename
        }
    }
    
    fragment HacktivityList on HacktivityItemConnection {
        pageInfo {
            endCursor
            hasNextPage
            __typename
        }
        edges {
            node {
                ... on HacktivityItemInterface {
                    id
                    databaseId: _id
                    __typename
                }
                __typename
            }
            ...HacktivityItem
            __typename
        }
        __typename
    }
    
    fragment HacktivityItem on HacktivityItemUnionEdge {
        node {
            ... on HacktivityItemInterface {
                id
                type: __typename
            }
            ... on Undisclosed {
                id
                ...HacktivityItemUndisclosed
                __typename
            }
            ... on Disclosed {
                id
                ...HacktivityItemDisclosed
                __typename
            }
            ... on HackerPublished {
                id
                ...HacktivityItemHackerPublished
                __typename
            }
            __typename
        }
        __typename
    }
    
    fragment HacktivityItemUndisclosed on Undisclosed {
        id
        votes {
            total_count
            __typename
        }
        upvoted: upvoted_by_current_user
        reporter {
            id
            username
            ...UserLinkWithMiniProfile
            __typename
        }
        team {
            handle
            name
            medium_profile_picture: profile_picture(size: medium)
            url
            id
            ...TeamLinkWithMiniProfile
            __typename
        }
        latest_disclosable_action
        latest_disclosable_activity_at
        requires_view_privilege
        total_awarded_amount
        currency
        __typename
    }
    
    fragment TeamLinkWithMiniProfile on Team {
        id
        handle
        name
        __typename
    }
    
    fragment UserLinkWithMiniProfile on User {
        id
        username
        __typename
    }
    
    fragment HacktivityItemDisclosed on Disclosed {
        id
        reporter {
            id
            username
            ...UserLinkWithMiniProfile
            __typename
        }
        votes {
            total_count
            __typename
        }
        upvoted: upvoted_by_current_user
        team {
            handle
            name
            medium_profile_picture: profile_picture(size: medium)
            url
            id
            ...TeamLinkWithMiniProfile
            __typename
        }
        report {
            id
            databaseId: _id
            title
            substate
            url
            created_at
            report_generated_content {
                id
                hacktivity_summary
                __typename
            }
            __typename
        }
        latest_disclosable_action
        latest_disclosable_activity_at
        total_awarded_amount
        severity_rating
        currency
        __typename
    }
    
    fragment HacktivityItemHackerPublished on HackerPublished {
        id
        reporter {
            id
            username
            ...UserLinkWithMiniProfile
            __typename
        }
        votes {
            total_count
            __typename
        }
        upvoted: upvoted_by_current_user
        team {
            id
            handle
            name
            medium_profile_picture: profile_picture(size: medium)
            url
            ...TeamLinkWithMiniProfile
            __typename
        }
        report {
            id
            url
            title
            substate
            __typename
        }
        latest_disclosable_activity_at
        severity_rating
        __typename
    }`

	variables := map[string]any{
		"count":           reports_count,
		"orderBy":         nil,
		"product_area":    "hacktivity",
		"product_feature": "overview",
		"secureOrderBy": map[string]any{
			"latest_disclosable_activity_at": map[string]any{
				"_direction": "DESC",
			},
		},
	}

	if disclosed_only {
		variables["where"] = map[string]any{
			"report": map[string]any{
				"disclosed_at": map[string]any{
					"_is_null": false,
				},
			},
		}
	}

	gql := map[string]any{
		"operationName": "HacktivityPageQuery",
		"query":         query,
		"variables":     variables,
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
