package teamcity

type BuildStatusResponse struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	State     string `json:"state"`
	Artifacts struct {
		Href string `json:"href"`
	} `json:"artifacts"`
	SnapshotDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID                  int    `json:"id"`
			BuildTypeID         string `json:"buildTypeId"`
			State               string `json:"state"`
			BranchName          string `json:"branchName"`
			Href                string `json:"href"`
			WebURL              string `json:"webUrl"`
			Customized          bool   `json:"customized"`
			MatrixConfiguration struct {
				Enabled bool `json:"enabled"`
			} `json:"matrixConfiguration"`
		} `json:"build"`
	} `json:"snapshot-dependencies"`
}

type TriggerBuildWithParametersResponse struct {
	ID          int    `json:"id"`
	BuildTypeID string `json:"buildTypeId"`
	State       string `json:"state"`
	Composite   bool   `json:"composite"`
	Href        string `json:"href"`
	WebURL      string `json:"webUrl"`
	BuildType   struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ProjectName string `json:"projectName"`
		ProjectID   string `json:"projectId"`
		Href        string `json:"href"`
		WebURL      string `json:"webUrl"`
	} `json:"buildType"`
	WaitReason string `json:"waitReason"`
	QueuedDate string `json:"queuedDate"`
	Triggered  struct {
		Type string `json:"type"`
		Date string `json:"date"`
		User struct {
			Username string `json:"username"`
			Name     string `json:"name"`
			ID       int    `json:"id"`
			Href     string `json:"href"`
		} `json:"user"`
	} `json:"triggered"`
	SnapshotDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID                  int    `json:"id"`
			BuildTypeID         string `json:"buildTypeId"`
			State               string `json:"state"`
			BranchName          string `json:"branchName"`
			DefaultBranch       bool   `json:"defaultBranch"`
			Href                string `json:"href"`
			WebURL              string `json:"webUrl"`
			MatrixConfiguration struct {
				Enabled bool `json:"enabled"`
			} `json:"matrixConfiguration"`
		} `json:"build"`
	} `json:"snapshot-dependencies"`
}

type ArtifactChildrenResponse struct {
	Count int `json:"count"`
	File  []struct {
		Name             string `json:"name"`
		Size             int    `json:"size"`
		ModificationTime string `json:"modificationTime"`
		Href             string `json:"href"`
		Content          struct {
			Href string `json:"href"`
		} `json:"content"`
	} `json:"file"`
}
