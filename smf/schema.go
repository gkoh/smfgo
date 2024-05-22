package smf

import (
	"encoding/xml"
)

const Header = "<?xml version='1.0'?>\n"
const DocType = "<!DOCTYPE service_bundle SYSTEM '/usr/share/lib/xml/dtd/service_bundle.dtd.1'>\n"

type ServiceBundle struct {
	XMLName xml.Name `xml:"service_bundle"`
	Type    string   `xml:"type,attr"`
	Name    string   `xml:"name,attr"`
	Service Service
}

type Service struct {
	XMLName xml.Name `xml:"service"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Version string   `xml:"version,attr"`
	*SingleInstance
	Instance  []Instance
	Stability Stability
	Template  Template
}

type SingleInstance struct {
	CreateDefaultInstance struct {
		XMLName xml.Name `xml:"create_default_instance"`
		Enabled bool     `xml:"enabled,attr"`
	}
	SingleInstance string `xml:"single_instance"`
	Dependencies   []Dependency
	Context        *MethodContext
	ExecMethods    []ExecMethod
	PropertyGroups []PropertyGroup
}

type Instance struct {
	XMLName        xml.Name `xml:"instance"`
	Name           string   `xml:"name,attr"`
	Enabled        bool     `xml:"enabled,attr"`
	Dependencies   []Dependency
	Context        *MethodContext
	ExecMethods    []ExecMethod
	PropertyGroups []PropertyGroup
}

type Stability struct {
	XMLName xml.Name `xml:"stability"`
	Value   string   `xml:"value,attr"`
}

type Dependency struct {
	XMLName     xml.Name `xml:"dependency"`
	Name        string   `xml:"name,attr"`
	Grouping    string   `xml:"grouping,attr"`
	RestartOn   string   `xml:"restart_on,attr"`
	Type        string   `xml:"type,attr"`
	ServiceFMRI ServiceFMRI
}

type ServiceFMRI struct {
	XMLName xml.Name `xml:"service_fmri"`
	Value   string   `xml:"value,attr"`
}

type MethodContext struct {
	XMLName     xml.Name `xml:"method_context"`
	Credential  *MethodCredential
	Environment *MethodEnvironment
}

type MethodCredential struct {
	XMLName         xml.Name `xml:"method_credential,omitempty"`
	User            string   `xml:"user,attr"`
	Group           string   `xml:"group,attr,omitempty"`
	SuppGroups      string   `xml:"supp_groups,attr,omitempty"`
	Privileges      string   `xml:"privileges,attr,omitempty"`
	LimitPrivileges string   `xml:"limit_privileges,attr,omitempty"`
}

type MethodEnvironment struct {
	XMLName xml.Name `xml:"method_environment"`
	EnvVar  []EnvVar
}

type EnvVar struct {
	XMLName xml.Name `xml:"envvar,omitempty"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type ExecMethod struct {
	XMLName        xml.Name `xml:"exec_method"`
	Name           string   `xml:"name,attr"`
	Type           string   `xml:"type,attr"`
	Exec           string   `xml:"exec,attr"`
	TimeOutSeconds uint     `xml:"timeout_seconds,attr"`
	Context        *MethodContext
}

type PropertyGroup struct {
	XMLName    xml.Name `xml:"property_group"`
	Name       string   `xml:"name,attr"`
	Type       string   `xml:"type,attr"`
	PropVals   []PropVal
	Properties []Property
}

type PropVal struct {
	XMLName xml.Name `xml:"propval,omitempty"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:"value,attr"`
}

type Property struct {
	XMLName xml.Name `xml:"property,omitempty"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
}

type Template struct {
	XMLName    xml.Name   `xml:"template"`
	CommonName CommonName `xml:"common_name"`
}

type CommonName struct {
	XMLName xml.Name `xml:"common_name"`
	LocText LocText  `xml:"loctext"`
}

type LocText struct {
	XMLName xml.Name `xml:"loctext"`
	Lang    string   `xml:"xml:lang,attr"`
	Text    string   `xml:",chardata"`
}

var DefaultDependencyNetwork Dependency = Dependency{
	Name: "network", Grouping: "require_all", RestartOn: "error", Type: "service", ServiceFMRI: ServiceFMRI{Value: "svc:/milestone/network:default"}}

var DefaultDependencyLocalFS Dependency = Dependency{
	Name: "filesystem-local", Grouping: "require_all", RestartOn: "none", Type: "service", ServiceFMRI: ServiceFMRI{Value: "svc:/system/filesystem/local:default"}}

var DefaultDependencyAutoFS Dependency = Dependency{
	Name: "autofs", Grouping: "optional_all", RestartOn: "none", Type: "service", ServiceFMRI: ServiceFMRI{Value: "svc:/system/filesystem/autofs:default"}}

var DefaultDependencies []Dependency = []Dependency{
	DefaultDependencyNetwork,
	DefaultDependencyLocalFS,
	DefaultDependencyAutoFS,
}

var DefaultPropertyGroups []PropertyGroup = []PropertyGroup{
	{
		Name: "startd",
		Type: "framework",
		PropVals: []PropVal{
			{
				Name:  "ignore_error",
				Type:  "astring",
				Value: "core,signal",
			},
		},
	},
	{
		Name: "general",
		Type: "framework",
		Properties: []Property{
			{
				Name: "action_authorization",
				Type: "astring",
			},
			{
				Name: "value_authorization",
				Type: "astring",
			},
		},
	},
}
