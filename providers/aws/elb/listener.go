package elb

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	. "github.com/tailwarden/komiser/models"
	. "github.com/tailwarden/komiser/providers"
)

func Listener(ctx context.Context, client ProviderClient) ([]Resource, error) {
	resources := make([]Resource, 0)

	var config elasticloadbalancingv2.DescribeLoadBalancersInput
	elbClient := elasticloadbalancingv2.NewFromConfig(*client.AWSClient)

	output, err := elbClient.DescribeLoadBalancers(ctx, &config)
	if err != nil {
		return resources, err
	}

	for _, loadbalancer := range output.LoadBalancers {
		loadbalancerArn := loadbalancer.LoadBalancerArn

		var configListeners elasticloadbalancingv2.DescribeListenersInput
		configListeners.LoadBalancerArn = loadbalancerArn
		elblisClient := elasticloadbalancingv2.NewFromConfig(*client.AWSClient)

		output, err := elblisClient.DescribeListeners(ctx, &configListeners)
		if err != nil {
			return resources, err

		}

		for _, listener := range output.Listeners {
			resourceArn := *listener.ListenerArn

			outputTags, err := elblisClient.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
				ResourceArns: []string{resourceArn},
			})
			if err != nil {
				return resources, err
			}

			tags := make([]Tag, 0)
			for _, tagDescription := range outputTags.TagDescriptions {
				for _, tag := range tagDescription.Tags {
					tags = append(tags, Tag{
						Key:   *tag.Key,
						Value: *tag.Value,
					})
				}
			}

			resources = append(resources, Resource{
				Provider:   "AWS",
				Account:    client.Name,
				Service:    "ELB Listener",
				ResourceId: resourceArn,
				Region:     client.AWSClient.Region,
				Name:       *listener.ListenerArn,
				Tags:       tags,
				FetchedAt:  time.Now(),
				Link:       fmt.Sprintf("https://%s.console.aws.amazon.com/ec2/home?region=%s#ELBListenerV2:listenerArn=%s", client.AWSClient.Region, client.AWSClient.Region, resourceArn),
			})
		}
	}

	log.WithFields(log.Fields{
		"provider":  "AWS",
		"account":   client.Name,
		"region":    client.AWSClient.Region,
		"service":   "ELB Listner",
		"resources": len(resources),
	}).Info("Fetched resources")

	return resources, nil
}
