package proxy

import (
	"fmt"
	"strconv"
	"testing"
)

func TestName(t *testing.T) {
	str := "[{\"channelName\":\"by-dev-rootcoord-dml_3_446582690160576429v2\",\"nodeID\":21},{\"channelName\":\"by-dev-rootcoord-dml_0_446582087845413659v0\",\"nodeID\":21},{\"channelName\":\"by-dev-rootcoord-dml_2_446582690160576429v1\",\"nodeID\":21},{\"channelName\":\"by-dev-rootcoord-dml_4_446582690160576429v3\",\"nodeID\":21},{\"channelName\":\"by-dev-rootcoord-dml_1_446582690160576429v0\",\"nodeID\":21}]"

	ret, err := strconv.Unquote(str)
	fmt.Println(ret)
	fmt.Println(err)

}
