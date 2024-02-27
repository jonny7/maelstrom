package dto

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
)

type Member struct {
	Name, Port, Addr, Tags, Status string
}

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

type Vortex struct {
	ID, Host, ConsumerBuffer, ResultBuffer, Workers, Jobs, RunningTime string
	Finished                                                           bool
}

func VortexToDTO(vortexes []maelstrom.Vortex) []Vortex {
	var runs []Vortex
	for _, v := range vortexes {
		now := getOptionalTime(v.EndTime)
		start := getOptionalTime(v.StartTime)
		runs = append(runs, Vortex{
			ID:             *v.Id,
			Host:           *v.Host,
			ConsumerBuffer: strconv.Itoa(v.ConsumerBuffer),
			ResultBuffer:   strconv.Itoa(v.ResultBuffer),
			Workers:        strconv.Itoa(v.Workers),
			Jobs:           strconv.Itoa(v.Jobs),
			RunningTime:    fmt.Sprintf("%v", now.Sub(start)),
			Finished:       timeToBool(v.EndTime),
		})
	}
	return runs
}

func timeToBool(t *int) bool {
	if t == nil {
		return false
	}
	return true
}

// @todo add tests here
func getOptionalTime(t *int) time.Time {
	if t == nil {
		return time.Now()
	}
	return time.Unix(int64(*t), 0)
}
