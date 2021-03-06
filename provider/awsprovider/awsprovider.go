package awsprovider

import (
	"fmt"

	"github.com/taskcluster/taskcluster-worker-runner/cfg"
	"github.com/taskcluster/taskcluster-worker-runner/protocol"
	"github.com/taskcluster/taskcluster-worker-runner/provider/provider"
	"github.com/taskcluster/taskcluster-worker-runner/run"
	"github.com/taskcluster/taskcluster-worker-runner/tc"
	tcclient "github.com/taskcluster/taskcluster/clients/client-go/v16"
	"github.com/taskcluster/taskcluster/clients/client-go/v16/tcworkermanager"
)

type AWSProvider struct {
	runnercfg                  *cfg.RunnerConfig
	workerManagerClientFactory tc.WorkerManagerClientFactory
	metadataService            MetadataService
	proto                      *protocol.Protocol
}

func (p *AWSProvider) ConfigureRun(state *run.State) error {
	userData, err := p.metadataService.queryUserData()
	if err != nil {
		return fmt.Errorf("Could not query user data: %v", err)
	}

	iid_string, iid_json, err := p.metadataService.queryInstanceIdentityDocument()
	if err != nil {
		return fmt.Errorf("Could not query instance identity document: %v", err)
	}

	instanceIdentityDocumentSignature, err := p.metadataService.queryMetadata("/dynamic/instance-identity/signature")
	if err != nil {
		return fmt.Errorf("Could not query signature for the instance identity document: %v", err)
	}

	state.RootURL = userData.RootURL
	state.WorkerLocation = map[string]string{
		"cloud":            "aws",
		"availabilityZone": iid_json.AvailabilityZone,
		"region":           iid_json.Region,
	}

	wm, err := p.workerManagerClientFactory(state.RootURL, nil)
	if err != nil {
		return fmt.Errorf("Could not create worker manager client: %v", err)
	}

	workerIdentityProofMap := map[string]interface{}{
		"document":  interface{}(iid_string),
		"signature": interface{}(instanceIdentityDocumentSignature),
	}

	err = provider.RegisterWorker(
		state,
		wm,
		userData.WorkerPoolId,
		userData.ProviderId,
		userData.WorkerGroup,
		iid_json.InstanceId,
		workerIdentityProofMap)
	if err != nil {
		return err
	}

	publicHostname, err := p.metadataService.queryMetadata("/meta-data/public-hostname")
	if err != nil {
		return err
	}

	publicIp, err := p.metadataService.queryMetadata("/meta-data/public-ipv4")
	if err != nil {
		return err
	}

	providerMetadata := map[string]string{
		"instance-id":       iid_json.InstanceId,
		"image":             iid_json.ImageId,
		"instance-type":     iid_json.InstanceType,
		"region":            iid_json.Region,
		"availability-zone": iid_json.AvailabilityZone,
		"local-ipv4":        iid_json.PrivateIp,
		"public-hostname":   publicHostname,
		"public-ipv4":       publicIp,
	}

	state.ProviderMetadata = providerMetadata

	return nil
}

func (p *AWSProvider) UseCachedRun(run *run.State) error {
	return nil
}

func (p *AWSProvider) SetProtocol(proto *protocol.Protocol) {
	p.proto = proto
}

func (p *AWSProvider) WorkerStarted() error {
	return nil
}

func (p *AWSProvider) WorkerFinished() error {
	return nil
}

func clientFactory(rootURL string, credentials *tcclient.Credentials) (tc.WorkerManager, error) {
	prov := tcworkermanager.New(credentials, rootURL)
	return prov, nil
}

func New(runnercfg *cfg.RunnerConfig) (provider.Provider, error) {
	return new(runnercfg, nil, nil)
}

func Usage() string {
	return `
The providerType "aws" is intended for workers provisioned with worker-manager
providers using providerType "aws".  It requires

	provider:
		providerType: aws

The TASKCLUSTER_WORKER_LOCATION of this provider has the following fields:

- cloud: aws
- region
- availabilityZone
`
}

// New takes its dependencies as optional arguments, allowing injection of fake dependencies for testing.
func new(
	runnercfg *cfg.RunnerConfig,
	workerManagerClientFactory tc.WorkerManagerClientFactory,
	metadataService MetadataService) (*AWSProvider, error) {

	if workerManagerClientFactory == nil {
		workerManagerClientFactory = clientFactory
	}

	if metadataService == nil {
		metadataService = &realMetadataService{}
	}

	return &AWSProvider{
		runnercfg:                  runnercfg,
		workerManagerClientFactory: workerManagerClientFactory,
		metadataService:            metadataService,
		proto:                      nil,
	}, nil
}
