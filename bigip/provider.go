/*
Original work from https://github.com/DealerDotCom/terraform-provider-bigip
Modifications Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file,You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Domain name/IP of the BigIP",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_HOST", nil),
			},
			"port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Management Port to connect to Bigip",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_PORT", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username with API access to the BigIP",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_USER", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The user's password. Leave empty if using token_value",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_PASSWORD", nil),
			},
			"token_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A token generated outside the provider, in place of password",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_TOKEN_VALUE", nil),
			},
			"token_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable to use an external authentication source (LDAP, TACACS, etc)",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_TOKEN_AUTH", nil),
			},
			"validate_certs_disable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enables TLS certificate check on BIG-IP",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_VERIFY_CERT_DISABLE", nil),
			},
			"trusted_cert_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Valid Trusted Certificate path",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_TRUSTED_CERT_PATH", nil),
			},
			"teem_disable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If this flag set to true,sending telemetry data to TEEM will be disabled",
				DefaultFunc: schema.EnvDefaultFunc("TEEM_DISABLE", false),
			},
			"login_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "tmos",
				Description: "Login reference for token authentication (see BIG-IP REST docs for details)",
				DefaultFunc: schema.EnvDefaultFunc("BIGIP_LOGIN_REF", nil),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bigip_ltm_datagroup":        dataSourceBigipLtmDataGroup(),
			"bigip_ltm_monitor":          dataSourceBigipLtmMonitor(),
			"bigip_ltm_irule":            dataSourceBigipLtmIrule(),
			"bigip_ssl_certificate":      dataSourceBigipSslCertificate(),
			"bigip_ltm_pool":             dataSourceBigipLtmPool(),
			"bigip_ltm_policy":           dataSourceBigipLtmPolicy(),
			"bigip_ltm_node":             dataSourceBigipLtmNode(),
			"bigip_vwan_config":          dataSourceBigipVwanconfig(),
			"bigip_waf_signatures":       dataSourceBigipWafSignatures(),
			"bigip_waf_policy":           dataSourceBigipWafPolicy(),
			"bigip_waf_pb_suggestions":   dataSourceBigipWafPb(),
			"bigip_waf_entity_url":       dataSourceBigipWafEntityUrl(),
			"bigip_waf_entity_parameter": dataSourceBigipWafEntityParameter(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"bigip_cm_device":                       resourceBigipCmDevice(),
			"bigip_cm_devicegroup":                  resourceBigipCmDevicegroup(),
			"bigip_net_route":                       resourceBigipNetRoute(),
			"bigip_net_selfip":                      resourceBigipNetSelfIP(),
			"bigip_net_vlan":                        resourceBigipNetVlan(),
			"bigip_ltm_irule":                       resourceBigipLtmIRule(),
			"bigip_ltm_datagroup":                   resourceBigipLtmDataGroup(),
			"bigip_ltm_monitor":                     resourceBigipLtmMonitor(),
			"bigip_ltm_node":                        resourceBigipLtmNode(),
			"bigip_ltm_pool":                        resourceBigipLtmPool(),
			"bigip_ltm_pool_attachment":             resourceBigipLtmPoolAttachment(),
			"bigip_ltm_policy":                      resourceBigipLtmPolicy(),
			"bigip_ltm_profile_fasthttp":            resourceBigipLtmProfileFasthttp(),
			"bigip_ltm_profile_fastl4":              resourceBigipLtmProfileFastl4(),
			"bigip_ltm_profile_http2":               resourceBigipLtmProfileHttp2(),
			"bigip_ltm_profile_httpcompress":        resourceBigipLtmProfileHttpcompress(),
			"bigip_ltm_profile_oneconnect":          resourceBigipLtmProfileOneconnect(),
			"bigip_ltm_profile_tcp":                 resourceBigipLtmProfileTcp(),
			"bigip_ltm_profile_ftp":                 resourceBigipLtmProfileFtp(),
			"bigip_ltm_profile_http":                resourceBigipLtmProfileHttp(),
			"bigip_ltm_persistence_profile_srcaddr": resourceBigipLtmPersistenceProfileSrcAddr(),
			"bigip_ltm_persistence_profile_dstaddr": resourceBigipLtmPersistenceProfileDstAddr(),
			"bigip_ltm_persistence_profile_ssl":     resourceBigipLtmPersistenceProfileSSL(),
			"bigip_ltm_persistence_profile_cookie":  resourceBigipLtmPersistenceProfileCookie(),
			"bigip_ltm_profile_server_ssl":          resourceBigipLtmProfileServerSsl(),
			"bigip_ltm_profile_client_ssl":          resourceBigipLtmProfileClientSsl(),
			"bigip_ltm_snat":                        resourceBigipLtmSnat(),
			"bigip_ltm_snatpool":                    resourceBigipLtmSnatpool(),
			"bigip_ltm_virtual_address":             resourceBigipLtmVirtualAddress(),
			"bigip_ltm_virtual_server":              resourceBigipLtmVirtualServer(),
			"bigip_sys_dns":                         resourceBigipSysDns(),
			"bigip_sys_iapp":                        resourceBigipSysIapp(),
			"bigip_sys_ntp":                         resourceBigipSysNtp(),
			"bigip_sys_provision":                   resourceBigipSysProvision(),
			"bigip_sys_snmp":                        resourceBigipSysSnmp(),
			"bigip_sys_snmp_traps":                  resourceBigipSysSnmpTraps(),
			"bigip_sys_bigiplicense":                resourceBigipSysBigiplicense(),
			"bigip_as3":                             resourceBigipAs3(),
			"bigip_do":                              resourceBigipDo(),
			"bigip_fast_template":                   resourceBigipFastTemplate(),
			"bigip_fast_application":                resourceBigipFastApp(),
			"bigip_fast_http_app":                   resourceBigipHttpFastApp(),
			"bigip_fast_https_app":                  resourceBigipFastHTTPSApp(),
			"bigip_fast_tcp_app":                    resourceBigipFastTcpApp(),
			"bigip_ssl_certificate":                 resourceBigipSslCertificate(),
			"bigip_ssl_key":                         resourceBigipSslKey(),
			"bigip_command":                         resourceBigipCommand(),
			"bigip_common_license_manage_bigiq":     resourceBigiqLicenseManage(),
			"bigip_bigiq_as3":                       resourceBigiqAs3(),
			"bigip_event_service_discovery":         resourceServiceDiscovery(),
			"bigip_traffic_selector":                resourceBigipTrafficselector(),
			"bigip_ipsec_policy":                    resourceBigipIpsecPolicy(),
			"bigip_net_tunnel":                      resourceBigipNetTunnel(),
			"bigip_net_ike_peer":                    resourceBigipNetIkePeer(),
			"bigip_ipsec_profile":                   resourceBigipIpsecProfile(),
			"bigip_waf_policy":                      resourceBigipAwafPolicy(),
		},
	}
	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}
	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	config := Config{
		Address:           d.Get("address").(string),
		Port:              d.Get("port").(string),
		Username:          d.Get("username").(string),
		Password:          d.Get("password").(string),
		Token:             d.Get("token_value").(string),
		CertVerifyDisable: d.Get("validate_certs_disable").(bool),
	}
	if d.Get("token_auth").(bool) {
		config.LoginReference = d.Get("login_ref").(string)
	}
	if !d.Get("validate_certs_disable").(bool) {
		if d.Get("trusted_cert_path").(string) == "" {
			return nil, fmt.Errorf("Valid Trust Certificate path not provided using :%+v ", "trusted_cert_path")
		}
		config.TrustedCertificate = d.Get("trusted_cert_path").(string)
	}
	cfg, err := config.Client()
	if err != nil {
		return cfg, err
	}
	if cfg != nil {
		cfg.UserAgent = fmt.Sprintf("Terraform/%s", terraformVersion)
		cfg.UserAgent += fmt.Sprintf("/terraform-provider-bigip/%s", getVersion())
		cfg.Teem = d.Get("teem_disable").(bool)
		// cfg.Transport.TLSClientConfig.InsecureSkipVerify = d.Get("validate_certs_disable").(bool)
	}
	return cfg, err
}

// Convert slice of strings to schema.TypeSet
func makeStringList(list *[]string) []interface{} {
	ilist := make([]interface{}, len(*list))
	for i, v := range *list {
		ilist[i] = v
	}
	return ilist
}

// Convert slice of strings to schema.Set
func makeStringSet(list *[]string) *schema.Set {
	ilist := make([]interface{}, len(*list))
	for i, v := range *list {
		ilist[i] = v
	}
	return schema.NewSet(schema.HashString, ilist)
}

// Convert schema.TypeList to a slice of strings
func listToStringSlice(s []interface{}) []string {
	list := make([]string, len(s))
	for i, v := range s {
		list[i] = v.(string)
	}
	return list
}

// Convert schema.Set to a slice of strings
func setToStringSlice(s *schema.Set) []string {
	list := make([]string, s.Len())
	for i, v := range s.List() {
		list[i] = v.(string)
	}
	return list
}

// Convert schema.Set to a slice of interface
func setToInterfaceSlice(s *schema.Set) []interface{} {
	list := make([]interface{}, s.Len())
	for i, v := range s.List() {
		list[i] = v.(string)
	}
	return list
}

// Copy map values into an object where map key == object field name (e.g. map[foo] == &{Foo: ...}
func mapEntity(d map[string]interface{}, obj interface{}) {
	val := reflect.ValueOf(obj).Elem()
	for field := range d {
		f := val.FieldByName(cases.Title(language.Und, cases.NoLower).String(field))
		if f.IsValid() {
			if f.Kind() == reflect.Slice {
				incoming := d[field].([]interface{})
				s := reflect.MakeSlice(f.Type(), len(incoming), len(incoming))
				for i := 0; i < len(incoming); i++ {
					if incoming[i] != nil {
						s.Index(i).Set(reflect.ValueOf(incoming[i]))
					}
				}
				f.Set(s)
			} else {
				f.Set(reflect.ValueOf(d[field]))
			}
		} else {
			f := val.FieldByName(cases.Title(language.Und, cases.NoLower).String(toCamelCase(field)))
			f.Set(reflect.ValueOf(d[field]))
		}
	}
}

// Convert Snakecase to Camelcase
func toCamelCase(str string) string {
	var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.ReplaceAll(s, "_", ""))
	})
}

// Convert Camelcase to Snakecase
func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func getVersion() string {
	return ProviderVersion
}
