package main

import (
	"fmt"
	"github.com/gkoh/smfgo/smf"
	"github.com/pterm/pterm"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: smfgo <service.xml>")
		os.Exit(1)
	}

	var filename string = os.Args[1]
	var err error

	bundle := smf.ServiceBundle{
		Type: "manifest",
		Service: smf.Service{
			Type: "service",
			Template: smf.Template{
				CommonName: smf.CommonName{
					LocText: smf.LocText{
						Lang: "C",
					},
				},
			},
			Stability: smf.Stability{
				Value: "Evolving",
			},
		},
	}

	// ask bundle name
	bundle.Name, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The name of the service bundle").WithDefaultValue("mybundle").Show()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// ask service name
	bundle.Service.Name, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The name of the service").WithDefaultValue("component/myservice").Show()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// ask service manifest version
	bundle.Service.Version, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The version of the service manifest").WithDefaultValue("1").Show()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// ask for human readable name
	bundle.Service.Template.CommonName.LocText.Text, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The human readable name of the service").WithDefaultValue("My service.").Show()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// ask multi instance
	multi_instance, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Does this service support multiple instances?").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		if !multi_instance {
			bundle.Service.SingleInstance = &smf.SingleInstance{}
		}
	}

	// ask config file
	config_path, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Full path to configuration file, blank if no configuration file").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if config_path != "" {
		pg := smf.PropertyGroup{Name: "config", Type: "application"}
		pg.PropVals = append(pg.PropVals, smf.PropVal{Name: "file", Type: "astring", Value: config_path})
		if bundle.Service.SingleInstance != nil {
			bundle.Service.SingleInstance.PropertyGroups = append(bundle.Service.SingleInstance.PropertyGroups, pg)
		} else {
			//bundle.Service.Instance.PropertyGroups = append(bundle.Service.Instance.PropertyGroups, pg)
		}
	}

	// ask methods for start/stop
	methods := [][]string{{"start", ""}, {"stop", ":kill"}}
	for _, method := range methods {
		var exec string = ""
		for exec == "" {
			exec, err = pterm.DefaultInteractiveTextInput.WithDefaultText(fmt.Sprintf("Full command to %s the service, use %%{config/file} to reference configuration file, eg. '/usr/bin/myservice --%s %%{config/file}'", method[0], method[0])).WithDefaultValue(method[1]).Show()
			if err != nil {
				log.Fatalf("%v", err)
			} else if exec != "" {
				if bundle.Service.SingleInstance != nil {
					em := smf.ExecMethod{Name: method[0], Type: "method", Exec: exec, TimeOutSeconds: 60}
					bundle.Service.SingleInstance.ExecMethods = append(bundle.Service.SingleInstance.ExecMethods, em)
				}
			}
		}
	}

	// ask envvars
	add_envvars, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Add environment variables for method execution?").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if add_envvars {
		for true {
			name, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Environment variable name (leave blank to continue)").Show()
			if err != nil {
				log.Fatalf("%v", err)
			} else if name == "" {
				break
			} else {
				value, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Environment variable value").Show()
				if err != nil {
					log.Fatalf("%v", err)
				} else {
					if bundle.Service.Context == nil {
						bundle.Service.Context = &smf.MethodContext{Environment: &smf.MethodEnvironment{}}
					}
					bundle.Service.Context.Environment.EnvVar = append(bundle.Service.Context.Environment.EnvVar, smf.EnvVar{Name: name, Value: value})
				}
			}
		}
	}

	// ask startd model (transient, child, contract)
	startd_model, err := pterm.DefaultInteractiveSelect.WithDefaultText("Select service managment method").WithOptions([]string{"transient", "child", "contract"}).WithDefaultOption("child").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		if bundle.Service.SingleInstance != nil {
			pv := smf.PropVal{Name: "duration", Type: "astring"}
			switch startd_model {
			case "transient":
				pv.Value = "transient"
			case "child":
				pv.Value = "child"
			case "contract":
				pv.Value = "contract"
			}
			pg := smf.PropertyGroup{Name: "startd", Type: "framework"}
			pg.PropVals = append(pg.PropVals, pv, smf.PropVal{Name: "ignore_error", Type: "astring", Value: "core,signal"})
			bundle.Service.SingleInstance.PropertyGroups = append(bundle.Service.SingleInstance.PropertyGroups, pg)
		}
	}

	// ask method credentials (user, group)
	user, err := pterm.DefaultInteractiveTextInput.WithDefaultText("User for execution methods").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if user != "" {
		if bundle.Service.Context == nil {
			bundle.Service.Context = &smf.MethodContext{}
		}
		if bundle.Service.Context.Credential == nil {
			bundle.Service.Context.Credential = &smf.MethodCredential{}
		}
		bundle.Service.Context.Credential.User = user
	}

	group, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Group for execution methods").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if group != "" {
		if bundle.Service.Context == nil {
			bundle.Service.Context = &smf.MethodContext{}
		}
		if bundle.Service.Context.Credential == nil {
			bundle.Service.Context.Credential = &smf.MethodCredential{}
		}
		bundle.Service.Context.Credential.Group = group
	}

	// depends network/localfs?
	deps := []string{"network", "local filesystems"}
	for _, dep := range deps {
		enable, err := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Does this service depend on %s being ready?", dep)).WithDefaultValue(true).Show()
		if err != nil {
			log.Fatalf("%v", err)
		} else {
			if bundle.Service.SingleInstance != nil && enable {
				switch dep {
				case "network":
					bundle.Service.SingleInstance.Dependencies = append(bundle.Service.SingleInstance.Dependencies, smf.DefaultDependencyNetwork)

				case "local filesystems":
					bundle.Service.SingleInstance.Dependencies = append(bundle.Service.SingleInstance.Dependencies, smf.DefaultDependencyLocalFS)
				}
			}
		}
	}

	// enable default?
	enable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Enable the service instance by default?").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else {
		if bundle.Service.SingleInstance != nil {
			bundle.Service.SingleInstance.CreateDefaultInstance.Enabled = enable
		}
	}

	// stability?
	bundle.Service.Stability.Value, err = pterm.DefaultInteractiveSelect.WithDefaultText("Select service stability").WithOptions([]string{"Standard", "Stable", "Evolving", "Unstable", "External", "Obsolete"}).WithDefaultOption("Evolving").Show()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	output, err := bundle.GenerateXML()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	err = os.WriteFile(filename, []byte(output), 0644)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
