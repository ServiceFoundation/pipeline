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

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/banzaicloud/pipeline/config"
	pkgCluster "github.com/banzaicloud/pipeline/pkg/cluster"
	modelOracle "github.com/banzaicloud/pipeline/pkg/providers/oracle/model"
	"github.com/banzaicloud/pipeline/secret"
	"github.com/banzaicloud/pipeline/utils"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const unknown = "unknown"

//TableName constants
const (
	TableNameClusters             = "clusters"
	TableNameAlibabaProperties    = "alibaba_cluster_properties"
	TableNameAlibabaNodePools     = "alibaba_node_pools"
	TableNameAmazonProperties     = "amazon_cluster_properties"
	TableNameAmazonNodePools      = "amazon_node_pools"
	TableNameAmazonEksProperties  = "amazon_eks_cluster_properties"
	TableNameAzureProperties      = "azure_cluster_properties"
	TableNameAzureNodePools       = "azure_node_pools"
	TableNameDummyProperties      = "dummy_cluster_properties"
	TableNameKubernetesProperties = "kubernetes_cluster_properties"
)

//ClusterModel describes the common cluster model
// Note: this model is being moved to github.com/banzaicloud/pipeline/pkg/model.ClusterModel
type ClusterModel struct {
	ID             uint   `gorm:"primary_key"`
	UID            string `gorm:"unique_index:idx_uid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `gorm:"unique_index:idx_unique_id" sql:"index"`
	Name           string     `gorm:"unique_index:idx_unique_id"`
	Location       string
	Cloud          string
	Distribution   string
	OrganizationId uint `gorm:"unique_index:idx_unique_id"`
	SecretId       string
	ConfigSecretId string
	SshSecretId    string
	Status         string
	RbacEnabled    bool
	Monitoring     bool
	Logging        bool
	StatusMessage  string `sql:"type:text;"`
	ACSK           ACSKClusterModel
	EC2            EC2ClusterModel
	AKS            AKSClusterModel
	EKS            EKSClusterModel
	Dummy          DummyClusterModel
	Kubernetes     KubernetesClusterModel
	OKE            modelOracle.Cluster
	CreatedBy      uint
}

// ACSKNodePoolModel describes Alibaba Cloud CS node groups model of a cluster
type ACSKNodePoolModel struct {
	ID                 uint `gorm:"primary_key"`
	CreatedAt          time.Time
	CreatedBy          uint
	ClusterModelId     uint   `gorm:"unique_index:idx_modelid_name"`
	Name               string `gorm:"unique_index:idx_modelid_name"`
	InstanceType       string
	SystemDiskCategory string
	SystemDiskSize     int
	Image              string
	Count              int
}

// ACSKClusterModel describes the Alibaba Cloud CS cluster model
type ACSKClusterModel struct {
	ClusterModelId           uint `gorm:"primary_key"`
	ClusterID                string
	RegionID                 string
	ZoneID                   string
	MasterInstanceType       string
	MasterSystemDiskCategory string
	MasterSystemDiskSize     int
	SNATEntry                bool
	SSHFlags                 bool
	NodePools                []*ACSKNodePoolModel `gorm:"foreignkey:ClusterModelId"`
}

//EC2ClusterModel describes the ec2 cluster model
type EC2ClusterModel struct {
	ClusterModelId     uint `gorm:"primary_key"`
	MasterInstanceType string
	MasterImage        string
	NodePools          []*AmazonNodePoolsModel `gorm:"foreignkey:ClusterModelId"`
}

//AmazonNodePoolsModel describes Amazon node groups model of a cluster
type AmazonNodePoolsModel struct {
	ID               uint `gorm:"primary_key"`
	CreatedAt        time.Time
	CreatedBy        uint
	ClusterModelId   uint   `gorm:"unique_index:idx_modelid_name"`
	Name             string `gorm:"unique_index:idx_modelid_name"`
	NodeSpotPrice    string
	Autoscaling      bool
	NodeMinCount     int
	NodeMaxCount     int
	Count            int
	NodeImage        string
	NodeInstanceType string
	Delete           bool `gorm:"-"`
}

//EKSClusterModel describes the ec2 cluster model
type EKSClusterModel struct {
	ClusterModelId uint `gorm:"primary_key"`

	//kubernetes "1.10"
	Version   string
	NodePools []*AmazonNodePoolsModel `gorm:"foreignkey:ClusterModelId"`
}

//AKSClusterModel describes the aks cluster model
type AKSClusterModel struct {
	ClusterModelId    uint `gorm:"primary_key"`
	ResourceGroup     string
	KubernetesVersion string
	NodePools         []*AKSNodePoolModel `gorm:"foreignkey:ClusterModelId"`
}

// AKSNodePoolModel describes AKS node pools model of a cluster
type AKSNodePoolModel struct {
	ID               uint `gorm:"primary_key"`
	CreatedAt        time.Time
	CreatedBy        uint
	ClusterModelId   uint   `gorm:"unique_index:idx_modelid_name"`
	Name             string `gorm:"unique_index:idx_modelid_name"`
	Autoscaling      bool
	NodeMinCount     int
	NodeMaxCount     int
	Count            int
	NodeInstanceType string
}

// DummyClusterModel describes the dummy cluster model
type DummyClusterModel struct {
	ClusterModelId    uint `gorm:"primary_key"`
	KubernetesVersion string
	NodeCount         int
}

//KubernetesClusterModel describes the build your own cluster model
type KubernetesClusterModel struct {
	ClusterModelId uint              `gorm:"primary_key"`
	Metadata       map[string]string `gorm:"-"`
	MetadataRaw    []byte            `gorm:"meta_data"`
}

func (cs *ClusterModel) BeforeCreate() (err error) {
	if cs.UID == "" {
		cs.UID = uuid.NewV4().String()
	}

	return
}

// BeforeSave converts the metadata into a json string in case of Kubernetes
func (cs *ClusterModel) BeforeSave() error {
	log.Info("Before save convert meta data")

	if cs.Cloud == pkgCluster.Kubernetes && cs.Kubernetes.MetadataRaw != nil && len(cs.Kubernetes.MetadataRaw) != 0 {
		out, err := json.Marshal(cs.Kubernetes.Metadata)
		if err != nil {
			log.Errorf("Error during convert map to json: %s", err.Error())
			return err
		}
		cs.Kubernetes.MetadataRaw = out
	}

	return nil
}

// AfterFind converts metadata json string into map in case of Kubernetes and sets NodeInstanceType and/or Location field(s)
// to unknown if they are empty
func (cs *ClusterModel) AfterFind() error {

	log.Info("After find convert metadata")

	if len(cs.Location) == 0 {
		cs.Location = unknown
	}

	if cs.Cloud == pkgCluster.Kubernetes && cs.Kubernetes.MetadataRaw != nil && len(cs.Kubernetes.MetadataRaw) != 0 {
		out, err := utils.ConvertJson2Map(cs.Kubernetes.MetadataRaw)
		if err != nil {
			log.Errorf("Error during convert json to map: %s", err.Error())
			return err
		}
		cs.Kubernetes.Metadata = out
	}

	return nil
}

//Save the cluster to DB
func (cs *ClusterModel) Save() error {
	db := config.DB()
	err := db.Save(&cs).Error
	if err != nil {
		return err
	}
	return nil
}

func (cs *ClusterModel) preDelete() {
	log := log.WithFields(logrus.Fields{"organization": cs.OrganizationId, "cluster": cs.ID})

	log.Info("Delete unused cluster secrets")
	if err := secret.Store.DeleteByClusterUID(cs.OrganizationId, cs.UID); err != nil {
		log.Errorf("Error during deleting secret: %s", err.Error())
	}
}

//Delete cluster from DB
func (cs *ClusterModel) Delete() error {

	log.Info("Delete config secret")
	cs.preDelete()

	db := config.DB()
	return db.Delete(&cs).Error
}

// TableName sets ClusterModel's table name
func (ClusterModel) TableName() string {
	return TableNameClusters
}

// String method prints formatted cluster fields
func (cs *ClusterModel) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Id: %d, Creation date: %s, Cloud: %s, Distribution: %s, ", cs.ID, cs.CreatedAt, cs.Cloud, cs.Distribution))

	switch cs.Distribution {
	case pkgCluster.EC2:
		// Write EC2 Master
		buffer.WriteString(fmt.Sprintf("Master instance type: %s, Master image: %s",
			cs.EC2.MasterInstanceType,
			cs.EC2.MasterImage))

		// Write EC2 Node
		for _, nodePool := range cs.EC2.NodePools {
			buffer.WriteString(fmt.Sprintf("NodePool Name: %s, Autoscaling: %v, InstanceType: %s, Spot price: %s, Min count: %d, Max count: %d, Count: %d, Node image: %s",
				nodePool.Name,
				nodePool.Autoscaling,
				nodePool.NodeInstanceType,
				nodePool.NodeSpotPrice,
				nodePool.NodeMinCount,
				nodePool.NodeMaxCount,
				nodePool.Count,
				nodePool.NodeImage))
		}
	case pkgCluster.EKS:
		// Write EKS Master
		buffer.WriteString(fmt.Sprintf("Master version: %s",
			cs.EKS.Version))

		// Write EKS Node
		for _, nodePool := range cs.EKS.NodePools {
			buffer.WriteString(fmt.Sprintf("NodePool Name: %s, Autoscaling: %v, InstanceType: %s, Spot price: %s, Min count: %d, Max count: %d, Count: %d, Node image: %s",
				nodePool.Name,
				nodePool.Autoscaling,
				nodePool.NodeInstanceType,
				nodePool.NodeSpotPrice,
				nodePool.NodeMinCount,
				nodePool.NodeMaxCount,
				nodePool.Count,
				nodePool.NodeImage))
		}
	case pkgCluster.AKS:
		// Write AKS
		buffer.WriteString(fmt.Sprintf("NodePools: %v, Kubernetes version: %s",
			cs.AKS.NodePools,
			cs.AKS.KubernetesVersion))
	case pkgCluster.Dummy:
		buffer.WriteString(fmt.Sprintf("Node count: %d, kubernetes version: %s",
			cs.Dummy.NodeCount,
			cs.Dummy.KubernetesVersion))
	case pkgCluster.Kubernetes:
		buffer.WriteString(fmt.Sprintf("Metadata: %#v", cs.Kubernetes.Metadata))
	}

	return buffer.String()
}

// TableName sets ACSKClusterModel's table name
func (ACSKClusterModel) TableName() string {
	return TableNameAlibabaProperties
}

// TableName sets ACSKNodePoolModel's table name
func (ACSKNodePoolModel) TableName() string {
	return TableNameAlibabaNodePools
}

// TableName sets AmazonClusterModel's table name
func (EC2ClusterModel) TableName() string {
	return TableNameAmazonProperties
}

// TableName sets AmazonNodePoolsModel's table name
func (AmazonNodePoolsModel) TableName() string {
	return TableNameAmazonNodePools
}

// TableName sets EKSClusterModel's table name
func (EKSClusterModel) TableName() string {
	return TableNameAmazonEksProperties
}

// TableName sets AzureClusterModel's table name
func (AKSClusterModel) TableName() string {
	return TableNameAzureProperties
}

// TableName sets AzureNodePoolModel's table name
func (AKSNodePoolModel) TableName() string {
	return TableNameAzureNodePools
}

//TableName sets the DummyClusterModel's table name
func (DummyClusterModel) TableName() string {
	return TableNameDummyProperties
}

//TableName sets the KubernetesClusterModel's table name
func (KubernetesClusterModel) TableName() string {
	return TableNameKubernetesProperties
}

// AfterUpdate removes marked node pool(s)
func (a *EKSClusterModel) AfterUpdate(scope *gorm.Scope) error {
	log.Info("Remove node pools marked for deletion")

	for _, nodePoolModel := range a.NodePools {
		if nodePoolModel.Delete {
			err := scope.DB().Delete(nodePoolModel).Error

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AfterUpdate removes marked node pool(s)
func (a *EC2ClusterModel) AfterUpdate(scope *gorm.Scope) error {
	log.Info("Remove node pools marked for deletion")

	for _, nodePoolModel := range a.NodePools {
		if nodePoolModel.Delete {
			err := scope.DB().Delete(nodePoolModel).Error

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateStatus updates the model's status and status message in database
func (cs *ClusterModel) UpdateStatus(status, statusMessage string) error {
	cs.Status = status
	cs.StatusMessage = statusMessage
	return cs.Save()
}

// UpdateConfigSecret updates the model's config secret id in database
func (cs *ClusterModel) UpdateConfigSecret(configSecretId string) error {
	cs.ConfigSecretId = configSecretId
	return cs.Save()
}

// UpdateSshSecret updates the model's ssh secret id in database
func (cs *ClusterModel) UpdateSshSecret(sshSecretId string) error {
	cs.SshSecretId = sshSecretId
	return cs.Save()
}
