package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	buildkiteAuthToken string
	buildkiteURL       string
)

// https://mholt.github.io/json-to-go/ is awesome
type VSTSPayload struct {
	SubscriptionID string `json:"subscriptionId"`
	NotificationID int    `json:"notificationId"`
	ID             string `json:"id"`
	EventType      string `json:"eventType"`
	PublisherID    string `json:"publisherId"`
	Message        struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"message"`
	DetailedMessage struct {
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
	} `json:"detailedMessage"`
	Resource struct {
		Commits []struct {
			CommitID string `json:"commitId"`
			Author   struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"author"`
			Committer struct {
				Name  string    `json:"name"`
				Email string    `json:"email"`
				Date  time.Time `json:"date"`
			} `json:"committer"`
			Comment string `json:"comment"`
			URL     string `json:"url"`
		} `json:"commits"`
		RefUpdates []struct {
			Name        string `json:"name"`
			OldObjectID string `json:"oldObjectId"`
			NewObjectID string `json:"newObjectId"`
		} `json:"refUpdates"`
		Repository struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Project struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				URL        string `json:"url"`
				State      string `json:"state"`
				Visibility string `json:"visibility"`
			} `json:"project"`
			DefaultBranch string `json:"defaultBranch"`
			RemoteURL     string `json:"remoteUrl"`
		} `json:"repository"`
		PushedBy struct {
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
			UniqueName  string `json:"uniqueName"`
		} `json:"pushedBy"`
		PushID int       `json:"pushId"`
		Date   time.Time `json:"date"`
		URL    string    `json:"url"`
	} `json:"resource"`
	ResourceVersion    string `json:"resourceVersion"`
	ResourceContainers struct {
		Collection struct {
			ID string `json:"id"`
		} `json:"collection"`
		Account struct {
			ID string `json:"id"`
		} `json:"account"`
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	} `json:"resourceContainers"`
	CreatedDate time.Time `json:"createdDate"`
}

type BuildkitePayload struct {
	Commit  string `json:"commit"`
	Branch  string `json:"branch"`
	Message string `json:"message"`
	Author  struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Env struct {
		MYENVVAR string `json:"MY_ENV_VAR"`
	} `json:"env"`
	MetaData struct {
		SomeBuildData  string `json:"some build data"`
		OtherBuildData bool   `json:"other build data"`
	} `json:"meta_data"`
}

type BuildkiteCreateBuildResponse struct {
	ID      string      `json:"id"`
	URL     string      `json:"url"`
	WebURL  string      `json:"web_url"`
	Number  int         `json:"number"`
	State   string      `json:"state"`
	Blocked bool        `json:"blocked"`
	Message string      `json:"message"`
	Commit  string      `json:"commit"`
	Branch  string      `json:"branch"`
	Tag     interface{} `json:"tag"`
	Env     struct {
		MYENVVAR string `json:"MY_ENV_VAR"`
	} `json:"env"`
	Source  string `json:"source"`
	Creator struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		AvatarURL string    `json:"avatar_url"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"creator"`
	CreatedAt   time.Time   `json:"created_at"`
	ScheduledAt time.Time   `json:"scheduled_at"`
	StartedAt   interface{} `json:"started_at"`
	FinishedAt  interface{} `json:"finished_at"`
	MetaData    struct {
		SomeBuildData  string `json:"some build data"`
		OtherBuildData string `json:"other build data"`
	} `json:"meta_data"`
	PullRequest interface{} `json:"pull_request"`
	Pipeline    struct {
		ID                              string      `json:"id"`
		URL                             string      `json:"url"`
		WebURL                          string      `json:"web_url"`
		Name                            string      `json:"name"`
		Description                     string      `json:"description"`
		Slug                            string      `json:"slug"`
		Repository                      string      `json:"repository"`
		BranchConfiguration             interface{} `json:"branch_configuration"`
		DefaultBranch                   string      `json:"default_branch"`
		SkipQueuedBranchBuilds          bool        `json:"skip_queued_branch_builds"`
		SkipQueuedBranchBuildsFilter    interface{} `json:"skip_queued_branch_builds_filter"`
		CancelRunningBranchBuilds       bool        `json:"cancel_running_branch_builds"`
		CancelRunningBranchBuildsFilter interface{} `json:"cancel_running_branch_builds_filter"`
		Provider                        struct {
			ID       string `json:"id"`
			Settings struct {
			} `json:"settings"`
		} `json:"provider"`
		BuildsURL string    `json:"builds_url"`
		BadgeURL  string    `json:"badge_url"`
		CreatedAt time.Time `json:"created_at"`
		Env       struct {
		} `json:"env"`
		ScheduledBuildsCount int `json:"scheduled_builds_count"`
		RunningBuildsCount   int `json:"running_builds_count"`
		ScheduledJobsCount   int `json:"scheduled_jobs_count"`
		RunningJobsCount     int `json:"running_jobs_count"`
		WaitingJobsCount     int `json:"waiting_jobs_count"`
		Steps                []struct {
			Type                string `json:"type"`
			Name                string `json:"name"`
			Command             string `json:"command"`
			ArtifactPaths       string `json:"artifact_paths"`
			BranchConfiguration string `json:"branch_configuration"`
			Env                 struct {
			} `json:"env"`
			TimeoutInMinutes interface{} `json:"timeout_in_minutes"`
			AgentQueryRules  []string    `json:"agent_query_rules"`
			Concurrency      interface{} `json:"concurrency"`
			Parallelism      interface{} `json:"parallelism"`
		} `json:"steps"`
	} `json:"pipeline"`
	Jobs []struct {
		ID              string      `json:"id"`
		Type            string      `json:"type"`
		Name            string      `json:"name"`
		AgentQueryRules []string    `json:"agent_query_rules"`
		State           string      `json:"state"`
		BuildURL        string      `json:"build_url"`
		WebURL          string      `json:"web_url"`
		LogURL          string      `json:"log_url"`
		RawLogURL       string      `json:"raw_log_url"`
		ArtifactsURL    string      `json:"artifacts_url"`
		Command         string      `json:"command"`
		ExitStatus      interface{} `json:"exit_status"`
		ArtifactPaths   string      `json:"artifact_paths"`
		Agent           interface{} `json:"agent"`
		CreatedAt       time.Time   `json:"created_at"`
		ScheduledAt     time.Time   `json:"scheduled_at"`
		StartedAt       interface{} `json:"started_at"`
		FinishedAt      interface{} `json:"finished_at"`
		Retried         bool        `json:"retried"`
		RetriedInJobID  interface{} `json:"retried_in_job_id"`
		RetriesCount    interface{} `json:"retries_count"`
	} `json:"jobs"`
}

func vstsHandler(resp http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Error(resp, "This endpoint doesn't support this request type", http.StatusMethodNotAllowed)
		return
	}

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	payload := VSTSPayload{}
	err = json.Unmarshal([]byte(reqBody), &payload)
	if err != nil {
		panic(err)
	}

	if payload.EventType != "git.push" {
		log.Printf("Unsupported event type: %s\n", payload.EventType)
		http.Error(resp, "Event not supported: "+payload.EventType, http.StatusBadRequest)
		return
	}

	buildkitePayload := BuildkitePayload{}

	// RefUpdates.Name can be:
	// branch push: /refs/heads/<branch-name>
	// tag push: /refs/tags/<tag-name>
	refsUpdatesArr := strings.Split(payload.Resource.RefUpdates[0].Name, "/")
	gitObjectType := refsUpdatesArr[1]
	gitObjectName := refsUpdatesArr[2]

	if gitObjectType == "heads" {
		buildkitePayload.Branch = gitObjectName
		if payload.Resource.RefUpdates[0].NewObjectID != "0000000000000000000000000000000000000000" {
			buildkitePayload.Commit = payload.Resource.RefUpdates[0].NewObjectID
		} else {
			log.Println("Branch deletion, ignoring.")
			return
		}
	} else {
		log.Println("Tag pushed, ignoring.")
		return
	}

	var (
		authorName    string
		authorEmail   string
		commitMessage string
	)

	if len(payload.Resource.Commits) == 0 {
		// This can happen when a branch is pushed with no new commits to it
		log.Println("No commit data for branch. Using PushedBy data for build author information")
		log.Println(string(reqBody))
		authorName = payload.Resource.PushedBy.DisplayName
		authorEmail = payload.Resource.PushedBy.UniqueName
		commitMessage = payload.DetailedMessage.Markdown
	} else {
		authorName = payload.Resource.Commits[0].Author.Name
		authorEmail = payload.Resource.Commits[0].Author.Email
		commitMessage = payload.Resource.Commits[0].Comment
	}

	buildkitePayload.Author.Name = authorName
	buildkitePayload.Author.Email = authorEmail
	buildkitePayload.Message = commitMessage

	buildkitePayloadString, err := json.Marshal(buildkitePayload)

	if err != nil {
		log.Println("Couldn't create buildkite post body")
	}

	log.Printf("Firing off a build to buildkite: %s %s %s\n", buildkitePayload.Author.Name, buildkitePayload.Branch, buildkitePayload.Commit)

	buildkiteReq, err := http.NewRequest("POST", buildkiteURL, bytes.NewBuffer(buildkitePayloadString))

	buildkiteReq.Header.Set("Authorization", "Bearer "+buildkiteAuthToken)
	buildkiteReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	buildkiteResp, err := client.Do(buildkiteReq)
	if err != nil {
		panic(err)
	}
	defer buildkiteResp.Body.Close()
	body, err := ioutil.ReadAll(buildkiteResp.Body)
	if err != nil {
		panic(err)
	}
	response := BuildkiteCreateBuildResponse{}
	err = json.Unmarshal([]byte(body), &response)
	fmt.Println("Build created: ", response.Jobs[0].WebURL)

}

func setupServer() {
	http.HandleFunc("/", vstsHandler)
	http.ListenAndServe(":8080", nil)
}

func main() {
	buildkiteURL = os.Getenv("BUILDKITE_URL")
	if buildkiteURL == "" {
		log.Fatal("Please specify BUILDKITE_URL")
	}

	buildkiteAuthToken = os.Getenv("BUILDKITE_AUTH_TOKEN")
	if buildkiteAuthToken == "" {
		log.Fatal("Please specify BUILDKITE_AUTH_TOKEN")
	}
	setupServer()
}
