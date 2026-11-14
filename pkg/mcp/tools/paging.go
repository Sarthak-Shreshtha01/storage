package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// page is the limit and offset arguments of a tool that pages through the
// entries of one object.
type page struct {
	limit  int
	offset int
}

// parsePage reads the optional limit (1 to maxLimit, defaultLimit when not
// given) and offset (0 or more) arguments.
func parsePage(args map[string]any, defaultLimit, maxLimit int) (page, *mcp.CallToolResult) {
	p := page{limit: defaultLimit}
	limit, set, toolErr := intArg(args, "limit")
	if toolErr != nil {
		return p, toolErr
	}
	if set {
		if limit < 1 || limit > maxLimit {
			return p, toolError(ErrCodeInvalidArgument,
				fmt.Sprintf("limit must be between 1 and %d", maxLimit),
				map[string]any{"argument": "limit"})
		}
		p.limit = limit
	}
	offset, _, toolErr := intArg(args, "offset")
	if toolErr != nil {
		return p, toolErr
	}
	if offset < 0 {
		return p, toolError(ErrCodeInvalidArgument, "offset must not be negative", map[string]any{"argument": "offset"})
	}
	p.offset = offset
	return p, nil
}

// pageInfo tells the caller where a page is in the whole result.
type pageInfo struct {
	Total      int  `json:"total"`
	Returned   int  `json:"returned"`
	Truncated  bool `json:"truncated"`
	NextOffset *int `json:"next_offset,omitempty"`
}

// paginate returns the entries of p and where they are in items. The page is
// never nil, so that it is [] and not null in JSON.
func paginate[T any](items []T, p page) ([]T, pageInfo) {
	info := pageInfo{Total: len(items)}
	out := []T{}
	if p.offset < len(items) {
		end := min(p.offset+p.limit, len(items))
		out = items[p.offset:end]
		if end < len(items) {
			info.Truncated = true
			next := end
			info.NextOffset = &next
		}
	}
	info.Returned = len(out)
	return out, info
}

// listLimit reads the optional limit argument of a tool that lists objects
// and pages with continue: 1 to pageSize, pageSize when not given. A smaller
// page is the way out when a page of pageSize is above the size cap.
func listLimit(args map[string]any, pageSize int) (int64, *mcp.CallToolResult) {
	limit, set, toolErr := intArg(args, "limit")
	if toolErr != nil {
		return 0, toolErr
	}
	if !set {
		return int64(pageSize), nil
	}
	if limit < 1 || limit > pageSize {
		return 0, toolError(ErrCodeInvalidArgument,
			fmt.Sprintf("limit must be between 1 and %d", pageSize),
			map[string]any{"argument": "limit"})
	}
	return int64(limit), nil
}

// listLimitOption is the limit argument of a tool that lists objects.
func listLimitOption(what string, pageSize int) mcp.ToolOption {
	return mcp.WithNumber("limit",
		mcp.Description(fmt.Sprintf("Maximum number of %s per page (optional, default and maximum %d); lower it if a page is too large, then page with continue", what, pageSize)),
	)
}
