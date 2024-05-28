package main

import (
	"fmt"
	"github.com/gkoh/smfgo/smf"
	"github.com/pterm/pterm"
	"log"
	"os"
)

func askInstanceCore(instance *smf.InstanceCore) {
	// ask config file
	configPath, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Full path to configuration file, blank if no configuration file").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if configPath != "" {
		pg := smf.PropertyGroup{Name: "config", Type: "application"}
		pg.PropVals = append(pg.PropVals, smf.PropVal{Name: "file", Type: "astring", Value: configPath})
		instance.PropertyGroups = append(instance.PropertyGroups, pg)
	}

	// ask methods for start/stop
	methods := [][]string{{"start", ""}, {"stop", ":kill"}}
	for _, method := range methods {
		var exec string
		for exec == "" {
			exec, err = pterm.DefaultInteractiveTextInput.WithDefaultText(fmt.Sprintf("Full command to %s the service, use %%{config/file} to reference configuration file, eg. '/usr/bin/myservice --%s %%{config/file}'", method[0], method[0])).WithDefaultValue(method[1]).Show()
			if err != nil {
				log.Fatalf("%v", err)
			} else if exec != "" {
				em := smf.ExecMethod{Name: method[0], Type: "method", Exec: exec, TimeOutSeconds: 60}
				instance.ExecMethods = append(instance.ExecMethods, em)
			}
		}
	}

	// ask envvars
	addEnvvars, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Add environment variables for method execution?").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if addEnvvars {
		for true {
			name, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Environment variable name (leave blank to continue)").Show()
			if err != nil {
				log.Fatalf("%v", err)
			} else if name == "" {
				break
			}

			value, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Environment variable value").Show()
			if err != nil {
				log.Fatalf("%v", err)
			}
			if instance.Context == nil {
				instance.Context = &smf.MethodContext{Environment: &smf.MethodEnvironment{}}
			}
			instance.Context.Environment.EnvVar = append(instance.Context.Environment.EnvVar, smf.EnvVar{Name: name, Value: value})
		}
	}

	// ask startd model (transient, child, contract)
	startdModel, err := pterm.DefaultInteractiveSelect.WithDefaultText("Select service management method").WithOptions([]string{"transient", "child", "contract"}).WithDefaultOption("child").Show()
	if err != nil {
		log.Fatalf("%v", err)
	}

	pv := smf.PropVal{Name: "duration", Type: "astring"}
	switch startdModel {
	case "transient":
		pv.Value = "transient"
	case "child":
		pv.Value = "child"
	case "contract":
		pv.Value = "contract"
	}
	pg := smf.PropertyGroup{Name: "startd", Type: "framework"}
	pg.PropVals = append(pg.PropVals, pv, smf.PropVal{Name: "ignore_error", Type: "astring", Value: "core,signal"})
	instance.PropertyGroups = append(instance.PropertyGroups, pg)

	// ask method credentials (user, group)
	user, err := pterm.DefaultInteractiveTextInput.WithDefaultText("User for execution methods").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if user != "" {
		if instance.Context == nil {
			instance.Context = &smf.MethodContext{}
		}
		if instance.Context.Credential == nil {
			instance.Context.Credential = &smf.MethodCredential{}
		}
		instance.Context.Credential.User = user
	}

	group, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Group for execution methods").Show()
	if err != nil {
		log.Fatalf("%v", err)
	} else if group != "" {
		if instance.Context == nil {
			instance.Context = &smf.MethodContext{}
		}
		if instance.Context.Credential == nil {
			instance.Context.Credential = &smf.MethodCredential{}
		}
		instance.Context.Credential.Group = group
	}

	// depends network/localfs?
	deps := []string{"network", "local filesystems"}
	for _, dep := range deps {
		enable, err := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Does this service depend on %s being ready?", dep)).WithDefaultValue(true).Show()
		if err != nil {
			log.Fatalf("%v", err)
		}

		if enable {
			switch dep {
			case "network":
				instance.Dependencies = append(instance.Dependencies, smf.DependencyLoopback, smf.DependencyNetwork)

			case "local filesystems":
				instance.Dependencies = append(instance.Dependencies, smf.DependencyLocalFS)
			}
		}
	}
}

func askServiceBundle() (smf.ServiceBundle, error) {
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
		return bundle, err
	}

	// ask service name
	bundle.Service.Name, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The name of the service").WithDefaultValue("component/myservice").Show()
	if err != nil {
		return bundle, err
	}

	// ask service manifest version
	bundle.Service.Version, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The version of the service manifest").WithDefaultValue("1").Show()
	if err != nil {
		return bundle, err
	}

	// ask for human readable name
	bundle.Service.Template.CommonName.LocText.Text, err = pterm.DefaultInteractiveTextInput.WithDefaultText("The human readable name of the service").WithDefaultValue("My service.").Show()
	if err != nil {
		return bundle, err
	}

	// ask multi instance
	multiInstance, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Does this service support multiple instances?").Show()
	if err != nil {
		return bundle, err
	}

	if !multiInstance {
		bundle.Service.SingleInstance = &smf.SingleInstance{}
		// enable by default?
		enable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Enable the service instance by default?").Show()
		if err != nil {
			return bundle, err
		}
		bundle.Service.SingleInstance.CreateDefaultInstance.Enabled = enable
		// complete remaining single instance configuration
		askInstanceCore(&bundle.Service.SingleInstance.InstanceCore)
	} else {
		for i := 0; ; i++ {
			// ask instance name, blank to exit, need at least one
			name, err := pterm.DefaultInteractiveTextInput.WithDefaultText("The name of the instance name (blank to continue)").WithDefaultValue(fmt.Sprintf("myinstance%d", i)).Show()
			if err != nil {
				return bundle, err
			} else if name == "" {
				if len(bundle.Service.Instance) < 1 {
					// need at least one
					i--
					continue
				}
				// blank == no more instances
				break
			}

			instance := smf.Instance{Name: name}
			askInstanceCore(&instance.InstanceCore)

			// enable default?
			enable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Enable the service instance by default?").Show()
			if err != nil {
				return bundle, err
			}

			instance.Enabled = enable
			bundle.Service.Instance = append(bundle.Service.Instance, instance)
		}
	}

	// stability?
	bundle.Service.Stability.Value, err = pterm.DefaultInteractiveSelect.WithDefaultText("Select service stability").WithOptions([]string{"Standard", "Stable", "Evolving", "Unstable", "External", "Obsolete"}).WithDefaultOption("Evolving").Show()
	if err != nil {
		return bundle, err
	}

	return bundle, nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: smfgo <service.xml>")
	}

	filename := os.Args[1]

	bundle, err := askServiceBundle()
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
