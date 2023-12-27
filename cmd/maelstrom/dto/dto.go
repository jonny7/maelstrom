package dto

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/serf/serf"
)

type Member struct {
	Name, Port, Addr, Tags, Status string
}

// @todo organize all these better
func MemberDTO(nodes []serf.Member) []Member {
	var members []Member
	for _, n := range nodes {
		members = append(members, Member{
			Name: n.Name,
			Port: strconv.Itoa(int(n.Port)),
			Addr: n.Addr.String(),
			Tags: func(m map[string]string) string {
				var out string
				for k, v := range m {
					out += fmt.Sprintf("%s:%s", k, v)
				}
				return out
			}(n.Tags),
			Status: n.Status.String(),
		})
	}
	return members
}
