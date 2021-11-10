package mixtape

const ProjectsRedisKey = "mixtape:projects"
const NextProjectIdRedisKey = "mixtape:next_project_id"

type Project struct {
	Id      uint64   `json:"id"`
	Name    string   `json:"Name"`
	Title   string   `json:"title"`
	Mixtape string   `json:"mixtape"`
	Blurb   string   `json:"blurb"`
	Channel string   `json:"channel"`
	Hosts   []string `json:"hosts,omitempty"`
}
