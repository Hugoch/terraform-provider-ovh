package ovh

import (
	"fmt"
	"log"
	"os"
	"os/user"

	ini "gopkg.in/ini.v1"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a schema.Provider for OVH.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_ENDPOINT", nil),
				Description: descriptions["endpoint"],
			},
			"application_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_APPLICATION_KEY", ""),
				Description: descriptions["application_key"],
			},
			"application_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_APPLICATION_SECRET", ""),
				Description: descriptions["application_secret"],
			},
			"consumer_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_CONSUMER_KEY", ""),
				Description: descriptions["consumer_key"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"ovh_cloud_region":               dataSourcePublicCloudRegion(),
			"ovh_cloud_regions":              dataSourcePublicCloudRegions(),
			"ovh_domain_zone":                dataSourceDomainZone(),
			"ovh_iploadbalancing":            dataSourceIpLoadbalancing(),
			"ovh_me_paymentmean_bankaccount": dataSourceMePaymentmeanBankaccount(),
			"ovh_me_paymentmean_creditcard":  dataSourceMePaymentmeanCreditcard(),

			// Legacy naming schema (publiccloud)
			"ovh_publiccloud_region": deprecated(dataSourcePublicCloudRegion(),
				"Use ovh_cloud_region data source instead"),
			"ovh_publiccloud_regions": deprecated(dataSourcePublicCloudRegions(),
				"Use ovh_cloud_regions data source instead"),
		},

		ResourcesMap: map[string]*schema.Resource{
			"ovh_iploadbalancing_tcp_farm":        resourceIpLoadbalancingTcpFarm(),
			"ovh_iploadbalancing_tcp_farm_server": resourceIpLoadbalancingTcpFarmServer(),
			"ovh_iploadbalancing_tcp_frontend":    resourceIpLoadbalancingTcpFrontend(),
			"ovh_iploadbalancing_http_route":      resourceIPLoadbalancingRouteHTTP(),
			"ovh_iploadbalancing_http_route_rule": resourceIPLoadbalancingRouteHTTPRule(),
			"ovh_iploadbalancing_refresh":         resourceIPLoadbalancingRefresh(),
			"ovh_domain_zone_record":              resourceOvhDomainZoneRecord(),
			"ovh_domain_zone_redirection":         resourceOvhDomainZoneRedirection(),
			"ovh_ip_reverse":                      resourceOvhIpReverse(),
			"ovh_cloud_network_private":           resourcePublicCloudPrivateNetwork(),
			"ovh_cloud_network_private_subnet":    resourcePublicCloudPrivateNetworkSubnet(),
			"ovh_cloud_user":                      resourcePublicCloudUser(),
			"ovh_vrack_cloudproject":              resourceVRackPublicCloudAttachment(),

			// Legacy naming schema (publiccloud)
			"ovh_publiccloud_private_network": deprecated(resourcePublicCloudPrivateNetwork(),
				"Use ovh_cloud_network_private resource instead"),
			"ovh_publiccloud_private_network_subnet": deprecated(resourcePublicCloudPrivateNetworkSubnet(),
				"Use ovh_cloud_network_private_subnet resource instead"),
			"ovh_publiccloud_user": deprecated(resourcePublicCloudUser(),
				"Use ovh_cloud_user resource instead"),
			"ovh_vrack_publiccloud_attachment": deprecated(resourceVRackPublicCloudAttachment(),
				"Use ovh_vrack_cloudproject resource instead"),
		},

		ConfigureFunc: configureProvider,
	}
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"endpoint": "The OVH API endpoint to target (ex: \"ovh-eu\").",

		"application_key": "The OVH API Application Key.",

		"application_secret": "The OVH API Application Secret.",
		"consumer_key":       "The OVH API Consumer key.",
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	userHome, err := currentUserHome()
	if err != nil {
		log.Fatal(err)
	}
	config := Config{
		Endpoint: d.Get("endpoint").(string),
	}
	configFile := fmt.Sprintf("%s/.ovh.conf", userHome)
	if _, err := os.Stat(configFile); err == nil {
		c, err := ini.Load(configFile)
		if err != nil {
			return nil, err
		}

		section, err := c.GetSection(d.Get("endpoint").(string))
		if err != nil {
			return nil, err
		}
		config.ApplicationKey = section.Key("application_key").String()
		config.ApplicationSecret = section.Key("application_secret").String()
		config.ConsumerKey = section.Key("consumer_key").String()
	}
	if v, ok := d.GetOk("application_key"); ok {
		config.ApplicationKey = v.(string)
	}
	if v, ok := d.GetOk("application_secret"); ok {
		config.ApplicationSecret = v.(string)
	}
	if v, ok := d.GetOk("consumer_key"); ok {
		config.ConsumerKey = v.(string)
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func deprecated(r *schema.Resource, msg string) *schema.Resource {
	r.DeprecationMessage = msg
	return r
}

// currentUserHome attempts to get current user's home directory
func currentUserHome() (string, error) {
	userHome := ""
	usr, err := user.Current()
	if err != nil {
		// Fallback by trying to read $HOME
		userHome = os.Getenv("HOME")
		if userHome != "" {
			err = nil
		}
	} else {
		userHome = usr.HomeDir
	}
	return userHome, nil
}
