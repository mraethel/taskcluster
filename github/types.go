// This source code file is AUTO-GENERATED by github.com/taskcluster/jsonschema2go

package github

import (
	"encoding/json"
	"errors"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

type (
	// A paginated list of builds
	//
	// See http://schemas.taskcluster.net/github/v1/build-list.json#
	Builds1 struct {

		// A simple list of builds.
		//
		// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds
		Builds []struct {

			// The initial creation time of the build. This is when it became pending.
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/created
			Created tcclient.Time `json:"created"`

			// The GitHub webhook deliveryId. Extracted from the header 'X-GitHub-Delivery'
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/eventId
			EventID string `json:"eventId"`

			// Type of Github event that triggered the build (i.e. push, pull_request.opened).
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/eventType
			EventType string `json:"eventType"`

			// Github organization associated with the build.
			//
			// Syntax:     ^([a-zA-Z0-9-_%]*)$
			// Min length: 1
			// Max length: 100
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/organization
			Organization string `json:"organization"`

			// Github repository associated with the build.
			//
			// Syntax:     ^([a-zA-Z0-9-_%]*)$
			// Min length: 1
			// Max length: 100
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/repository
			Repository string `json:"repository"`

			// Github revision associated with the build.
			//
			// Min length: 40
			// Max length: 40
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/sha
			Sha string `json:"sha"`

			// Github status associated with the build.
			//
			// Possible values:
			//   * "pending"
			//   * "success"
			//   * "error"
			//   * "failure"
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/state
			State string `json:"state"`

			// Taskcluster task-group associated with the build.
			//
			// Syntax:     ^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/taskGroupId
			TaskGroupID string `json:"taskGroupId"`

			// The last updated of the build. If it is done, this is when it finished.
			//
			// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/updated
			Updated tcclient.Time `json:"updated"`
		} `json:"builds"`

		// Passed back from Azure to allow us to page through long result sets.
		//
		// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/continuationToken
		ContinuationToken string `json:"continuationToken,omitempty"`
	}

	// Syntax:     ^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}$
	//
	// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/eventId/oneOf[0]
	Var json.RawMessage

	// Syntax:     Unknown
	//
	// See http://schemas.taskcluster.net/github/v1/build-list.json#/properties/builds/items/properties/eventId/oneOf[1]
	Var1 json.RawMessage
)

// MarshalJSON calls json.RawMessage method of the same name. Required since
// Var is of type json.RawMessage...
func (this *Var) MarshalJSON() ([]byte, error) {
	x := json.RawMessage(*this)
	return (&x).MarshalJSON()
}

// UnmarshalJSON is a copy of the json.RawMessage implementation.
func (this *Var) UnmarshalJSON(data []byte) error {
	if this == nil {
		return errors.New("Var: UnmarshalJSON on nil pointer")
	}
	*this = append((*this)[0:0], data...)
	return nil
}

// MarshalJSON calls json.RawMessage method of the same name. Required since
// Var1 is of type json.RawMessage...
func (this *Var1) MarshalJSON() ([]byte, error) {
	x := json.RawMessage(*this)
	return (&x).MarshalJSON()
}

// UnmarshalJSON is a copy of the json.RawMessage implementation.
func (this *Var1) UnmarshalJSON(data []byte) error {
	if this == nil {
		return errors.New("Var1: UnmarshalJSON on nil pointer")
	}
	*this = append((*this)[0:0], data...)
	return nil
}
