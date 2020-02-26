package kafka

import "knative.dev/streaming/pkg/apis/streaming/v1alpha1"

func TopicName(stream *v1alpha1.Stream) string {
	return stream.Name
}
