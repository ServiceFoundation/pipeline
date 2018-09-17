// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

// CreateBucketRequest describes create bucket request
type CreateBucketRequest struct {
	Cloud      string
	BucketName string
	Location   string
	SecretID   string
}

// FindBucketRequest describes a find bucket request
type FindBucketRequest struct {
	Cloud      string
	BucketName string
	Location   string
}

// Bucket describes a Bucket used for ARK backups
type Bucket struct {
	ID                  uint   `json:"id"`
	Name                string `json:"name"`
	Cloud               string `json:"cloud"`
	SecretID            string `json:"secretId"`
	Location            string `json:"location,omitempty"`
	Status              string `json:"status"`
	InUse               bool   `json:"inUse"`
	DeploymentID        uint   `json:"deploymentId,omitempty"`
	ClusterID           uint   `json:"clusterId,omitempty"`
	ClusterCloud        string `json:"clusterCloud,omitempty"`
	ClusterDistribution string `json:"clusterDistribution,omitempty"`
}

// DeleteBucketResponse describes a delete bucket response
type DeleteBucketResponse struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}
