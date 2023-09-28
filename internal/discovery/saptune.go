package discovery

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/saptune"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/pkg/utils"
)

const SaptuneDiscoveryID string = "saptune_discovery"
const SaptuneDiscoveryMinPeriod time.Duration = 1 * time.Second

type SaptuneDiscovery struct {
	id              string
	collectorClient collector.Client
	interval        time.Duration
}

type SaptuneDiscoveryPayload struct {
	PackageVersion   string      `json:"package_version"`
	SaptuneInstalled bool        `json:"saptune_installed"`
	Status           interface{} `json:"status"`
}

func NewSaptuneDiscovery(collectorClient collector.Client, config DiscoveriesConfig) Discovery {
	return SaptuneDiscovery{
		id:              SaptuneDiscoveryID,
		collectorClient: collectorClient,
		interval:        config.DiscoveriesPeriodsConfig.Saptune,
	}
}

func (d SaptuneDiscovery) GetID() string {
	return d.id
}

func (d SaptuneDiscovery) GetInterval() time.Duration {
	return d.interval
}

func (d SaptuneDiscovery) Discover() (string, error) {
	var saptunePayload SaptuneDiscoveryPayload

	saptuneRetriever, err := saptune.NewSaptune(utils.Executor{})
	switch {
	case err != nil:
		saptunePayload = SaptuneDiscoveryPayload{
			PackageVersion:   "",
			SaptuneInstalled: false,
			Status:           nil,
		}
	case !saptuneRetriever.IsJSONSupported:
		saptunePayload = SaptuneDiscoveryPayload{
			PackageVersion:   saptuneRetriever.Version,
			SaptuneInstalled: true,
			Status:           nil,
		}
	default:
		saptuneData, _ := saptuneRetriever.RunCommandJSON("status")
		unmarshalled := make(map[string]interface{})
		err = json.Unmarshal(saptuneData, &unmarshalled)
		if err != nil {
			log.Debugf("Error while unmarshalling saptune status: %s", err)
			saptunePayload = SaptuneDiscoveryPayload{
				PackageVersion:   saptuneRetriever.Version,
				SaptuneInstalled: true,
				Status:           nil,
			}
		} else {
			saptunePayload = SaptuneDiscoveryPayload{
				PackageVersion:   saptuneRetriever.Version,
				SaptuneInstalled: true,
				Status:           unmarshalled,
			}
		}
	}

	err = d.collectorClient.Publish(d.id, saptunePayload)
	if err != nil {
		log.Debugf("Error while sending saptune discovery to data collector: %s", err)
		return "", err
	}

	return "Saptune data discovery completed", nil
}