package dto

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jonny7/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/internal/core/commander"
)

type Member struct {
	Name, Port, Addr, Tags, Status string
}

func MemberDTO(nodes []agent.Member) []Member {
	var members []Member
	for _, n := range nodes {
		members = append(members, Member{
			Name: n.Name,
			Port: strconv.Itoa(n.Port),
			Addr: n.Addr,
			Tags: func(m map[string]string) string {
				var out string
				for k, v := range m {
					out += fmt.Sprintf("%s:%s", k, v)
				}
				return out
			}(n.Tags),
			Status: n.Status,
		})
	}
	return members
}

type Vortex struct {
	ID, Host, ConsumerBuffer, ResultBuffer, Workers, Jobs, RunningTime string
	Finished                                                           bool
}

func VortexToDTO(vortexes []commander.Vortex) []Vortex {
	var runs []Vortex
	for _, v := range vortexes {
		end := v.EndTime
		if !v.Finished() {
			end = time.Now()
		}
		runs = append(runs, Vortex{
			ID:             v.ID.String(),
			Host:           v.Host,
			ConsumerBuffer: strconv.Itoa(v.ConsumerBuffer),
			ResultBuffer:   strconv.Itoa(v.ResultBuffer),
			Workers:        strconv.Itoa(v.Workers),
			Jobs:           strconv.Itoa(v.Jobs),
			RunningTime:    fmt.Sprintf("%v", end.Sub(v.StartTime).Round(time.Second)),
			Finished:       v.Finished(),
		})
	}
	return runs
}
