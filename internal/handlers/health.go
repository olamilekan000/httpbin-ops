package handlers

import "net/http"

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

func HealthHandler(buildInfo *BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		resp := map[string]string{"status": "ok"}
		if buildInfo != nil {
			resp["version"] = buildInfo.Version
			resp["commit"] = buildInfo.Commit
			resp["build_time"] = buildInfo.BuildTime
		}
		writeJSONResponse(w, http.StatusOK, resp)
	}
}
