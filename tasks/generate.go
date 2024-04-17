//go:build generate
// +build generate

package tasks

//go:generate mockgen -source=tasks.go -package mock -typed -destination mock/mock.go
