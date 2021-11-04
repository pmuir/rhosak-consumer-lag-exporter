package exporters

import (
	"fmt"
	kafkainstanceclient "github.com/redhat-developer/app-services-sdk-go/kafkainstance/apiv1internal/client"
	"gopkg.in/errgo.v2/errors"
	"io"
	"text/template")

type GroupTopic struct {
	BootstrapServer string
	GroupId         string
	Topic string
	Value int32
}

type Group struct {
	BootstrapServer string
	GroupId string
	Value int32
}

type BootstrapServerConsumerGroup struct {
	kafkainstanceclient.ConsumerGroup
	BootstrapServer string
}

type ExportData struct {
	ConsumerGroups []BootstrapServerConsumerGroup
	GroupTopicSums []GroupTopic
	GroupSums []Group
	GroupMaxs []Group
}

const PrometheusTmpl = `# HELP kafka_consumergroup_group_lag Group offset lag of a partition
# TYPE kafka_consumergroup_group_lag gauge
{{- range .ConsumerGroups }}
{{- $BootstrapServer := .BootstrapServer}}
{{- range .Consumers }}
kafka_consumergroup_group_lag{bootstrap_server="{{ $BootstrapServer }}",group="{{ .GroupId }}",topic="{{ .Topic }}",partition="{{ .Partition }}",consumer_id="{{ fixConsumerId .MemberId }}",} {{ fixLag .Lag }}
{{- end }}
{{- end }}
# HELP kafka_consumergroup_group_topic_sum_lag Sum of group offset lag across topic partitions
# TYPE kafka_consumergroup_group_topic_sum_lag gauge
{{- range .GroupTopicSums }}
kafka_consumergroup_group_topic_sum_lag{bootstrap_server="{{ .BootstrapServer }}",group="{{ .GroupId }}",topic="{{ .Topic }}",} {{ fixLag .Value }}
{{- end }}
# HELP kafka_consumergroup_group_sum_lag Sum of group offset lag
# TYPE kafka_consumergroup_group_sum_lag gauge
{{- range .GroupSums }}
kafka_consumergroup_group_sum_lag{bootstrap_server="{{ .BootstrapServer }}",group="{{ .GroupId }}",} {{ fixLag .Value }}
{{- end }}
# HELP kafka_consumergroup_group_max_lag Max group offset lag
# TYPE kafka_consumergroup_group_max_lag gauge
{{- range .GroupMaxs }}
kafka_consumergroup_group_max_lag{bootstrap_server="{{ .BootstrapServer }}",group="{{ .GroupId }}",} {{ fixLag .Value }}
{{- end }}
`

func AsPrometheus(servers map[string][]kafkainstanceclient.ConsumerGroup, output io.Writer) error {
	funcMap := template.FuncMap{
		"fixLag": func(lag int32) string {
			return fmt.Sprintf("%.1f", float32(lag))
		},
		"fixConsumerId": func(consumerId *string) string {
			if consumerId == nil {
				return ""
			}
			return *consumerId
		},
	}

	tmpl, err := template.New("prometheus").Funcs(funcMap).Parse(PrometheusTmpl)
	if err != nil {
		return errors.Wrap(err)
	}
	data := ExportData{
	}

	for bootstrapServer, consumerGroups := range servers {
		for _, item := range consumerGroups {
			data.ConsumerGroups = append(data.ConsumerGroups, BootstrapServerConsumerGroup{
				item,
				bootstrapServer,
			})
			for _, consumer := range item.GetConsumers() {
				foundGroupTopicSum := false
				for i, groupTopicSum := range data.GroupTopicSums {
					if groupTopicSum.BootstrapServer == bootstrapServer && groupTopicSum.GroupId == consumer.GetGroupId() && groupTopicSum.Topic == consumer.GetTopic() {
						data.GroupTopicSums[i].Value += consumer.Lag
						foundGroupTopicSum = true
						break
					}
				}
				if !foundGroupTopicSum {
					data.GroupTopicSums = append(data.GroupTopicSums, GroupTopic{
						BootstrapServer: bootstrapServer,
						GroupId:         consumer.GetGroupId(),
						Topic:           consumer.GetTopic(),
						Value:           consumer.GetLag(),
					})
				}
				foundGroupSum := false
				for i, groupSum := range data.GroupSums {
					if groupSum.BootstrapServer == bootstrapServer && groupSum.GroupId == consumer.GetGroupId() {
						data.GroupSums[i].Value += consumer.Lag
						foundGroupSum = true
						break
					}
				}
				if !foundGroupSum {
					data.GroupSums = append(data.GroupSums, Group{
						BootstrapServer: bootstrapServer,
						GroupId:         consumer.GetGroupId(),
						Value:           consumer.GetLag(),
					})
				}
				foundGroupMax := false
				for i, groupMax := range data.GroupMaxs {
					if groupMax.GroupId == consumer.GetGroupId() {
						if groupMax.BootstrapServer == bootstrapServer && consumer.GetLag() > groupMax.Value {
							data.GroupMaxs[i].Value = consumer.GetLag()
							break
						}
						foundGroupMax = true
					}
				}
				if !foundGroupMax {
					data.GroupMaxs = append(data.GroupMaxs, Group{
						BootstrapServer: bootstrapServer,
						GroupId: consumer.GetGroupId(),
						Value:   consumer.GetLag(),
					})
				}
			}
		}
	}

	err = tmpl.Execute(output, data)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
