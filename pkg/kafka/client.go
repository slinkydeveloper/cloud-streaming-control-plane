package kafka

import "github.com/Shopify/sarama"

func MakeClusterAdminClient(clientID string, bootstrapServers []string) (sarama.ClusterAdmin, error) {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V1_1_0_0
	saramaConf.ClientID = clientID
	return sarama.NewClusterAdmin(bootstrapServers, saramaConf)
}
