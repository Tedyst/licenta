package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

const defaultPerPage = 10
const maxPerPage = 100

type PaginationResponse[T any] struct {
	TotalCount  int32 `json:"total_count"`
	TotalPages  int32 `json:"total_pages"`
	CurrentPage int32 `json:"current_page"`
	PerPage     int32 `json:"per_page"`
	Data        T     `json:"data"`
}

func NewPaginationResponse[T any](data T, totalCount, currentPage, perPage int32) PaginationResponse[T] {
	return PaginationResponse[T]{
		TotalCount:  totalCount,
		TotalPages:  totalCount / perPage,
		CurrentPage: currentPage,
		PerPage:     perPage,
		Data:        data,
	}
}

func GetPageAndLimit(c *fiber.Ctx) (int32, int32, error) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return 0, 0, err
	}
	if page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "0"))
	if err != nil {
		return 0, 0, err
	}
	if limit == 0 {
		limit = defaultPerPage
	}
	if limit > maxPerPage {
		limit = maxPerPage
	}
	return int32(page), int32(limit), nil
}

func getOffset(page, limit int32) int32 {
	return (page - 1) * limit
}
