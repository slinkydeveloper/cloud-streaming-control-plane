package kafka

import "github.com/Shopify/sarama"

var (
	HARDCODED_BOOTSTRAP_SERVER = "my-cluster-kafka-bootstrap.kafka.svc:9092"
)

func MakeClusterAdminClient(clientID string, bootstrapServers []string) (sarama.ClusterAdmin, error) {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V1_1_0_0
	saramaConf.ClientID = clientID
	return sarama.NewClusterAdmin(bootstrapServers, saramaConf)
}
