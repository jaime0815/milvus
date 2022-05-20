package tools

import "github.com/milvus-io/milvus/cmd/tools/mck"

func RunTool(args []string) bool {
	switch {
	case len(args) >= 2 && args[1] == mck.MckCmd:
		mck.RunMck(args)
	default:
		return false
	}
	return true
}
