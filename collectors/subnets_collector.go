package collectors

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	api "github.com/tomvil/phpipam_exporter/client"
)

type SubnetsCollector struct {
	apiClient       *api.Client
	IPv4SubnetsUsed *prometheus.Desc
	IPv4SubnetsFree *prometheus.Desc
	IPv6SubnetsUsed *prometheus.Desc
	IPv6SubnetsFree *prometheus.Desc
}

type Subnets struct {
	Data []struct {
		Subnet      string
		Mask        string
		Description string
		Custom_free string
		Usage       struct {
			Used      interface{}
			Maxhosts  interface{}
			Freehosts interface{}
		}
	}
}

type Sections struct {
	Data []struct {
		Id   string
		Name string
	}
}

func NewSubnetsCollector(apiclient *api.Client) *SubnetsCollector {
	prefix := "phpipam_subnets_"
	return &SubnetsCollector{
		apiClient:       apiclient,
		IPv4SubnetsUsed: prometheus.NewDesc(prefix+"ipv4_used", "Number of used IPv4 subnets in phpIPAM", []string{"section", "mask"}, nil),
		IPv4SubnetsFree: prometheus.NewDesc(prefix+"ipv4_free", "Number of free IPv4 subnets in phpIPAM", []string{"section", "mask"}, nil),
		IPv6SubnetsUsed: prometheus.NewDesc(prefix+"ipv6_used", "Number of used IPv6 subnets in phpIPAM", []string{"section", "mask"}, nil),
		IPv6SubnetsFree: prometheus.NewDesc(prefix+"ipv6_free", "Number of free IPv6 subnets in phpIPAM", []string{"section", "mask"}, nil),
	}
}

func (c *SubnetsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.IPv4SubnetsUsed
	ch <- c.IPv4SubnetsFree
	ch <- c.IPv6SubnetsUsed
	ch <- c.IPv6SubnetsFree
}

func (c *SubnetsCollector) Collect(ch chan<- prometheus.Metric) {
	var sections Sections
	var subnets Subnets

	err := c.apiClient.GetParsed("/sections", &sections)
	if err != nil {
		log.Errorln(err.Error())
	}

	for _, section := range sections.Data {
		err := c.apiClient.GetParsed("/sections/"+section.Id+"/subnets", &subnets)
		if err != nil {
			log.Errorln(err.Error())
			continue
		}

		subnets4_free := make(map[string]int)
		subnets4_used := make(map[string]int)
		subnets6_free := make(map[string]int)
		subnets6_used := make(map[string]int)

		for _, subnet := range subnets.Data {
			if subnet.Custom_free == "1" {
				if IsIPv6(subnet.Subnet) {
					subnets6_free[subnet.Mask] += 1
				} else {
					subnets4_free[subnet.Mask] += 1
				}
			} else {
				if IsIPv6(subnet.Subnet) {
					subnets6_used[subnet.Mask] += 1
				} else {
					subnets4_used[subnet.Mask] += 1
				}
			}
		}

		for prefix, free_num := range subnets4_free {
			ch <- prometheus.MustNewConstMetric(c.IPv4SubnetsFree, prometheus.GaugeValue, float64(free_num), section.Name, prefix)
		}

		for prefix, used_num := range subnets4_used {
			ch <- prometheus.MustNewConstMetric(c.IPv4SubnetsUsed, prometheus.GaugeValue, float64(used_num), section.Name, prefix)
		}

		for prefix, free_num := range subnets6_free {
			ch <- prometheus.MustNewConstMetric(c.IPv6SubnetsFree, prometheus.GaugeValue, float64(free_num), section.Name, prefix)
		}

		for prefix, used_num := range subnets6_used {
			ch <- prometheus.MustNewConstMetric(c.IPv6SubnetsUsed, prometheus.GaugeValue, float64(used_num), section.Name, prefix)
		}
	}
}

func IsIPv6(address string) bool {
	return strings.Contains(address, ":")
}
