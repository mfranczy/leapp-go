package api

import (
	"encoding/json"
	"net/http"
)

type portInspectParams struct {
	TargetHost  string `json:"target_host"`
	PortRange   string `json:"port_range"`
	ShallowScan bool   `json:"shallow_scan"`
}

func portInspect(rw http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	var params portInspectParams

	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		return nil, http.StatusBadRequest, newAPIError(err, errBadInput, "could not decode data sent by client")
	}

	d := map[string][]interface{}{
		"host": ChannelData(ObjValue{params.TargetHost}),
		"scan_options": ChannelData(map[string]interface{}{
			"shallow_scan": params.ShallowScan,
			"port_range":   params.PortRange,
			"force_nmap":   !params.ShallowScan,
		}),
	}

	actorInput, err := json.Marshal(d)
	if err != nil {
		return nil, http.StatusInternalServerError, newAPIError(err, errInternal, "could not build actor's input")
	}

	id := actorRunnerRegistry.Create("portscan", string(actorInput))

	s := actorRunnerRegistry.GetStatus(id, true)
	if err := checkSyncTaskStatus(s); err != nil {
		return nil, http.StatusInternalServerError, newAPIError(err, errInternal, "")
	}

	logExecutorError(req.Context(), s.Result)

	return parseExecutorResult(s.Result)
}
