# smfgo - SMF manifest file generator

A standalone SMF manifest file generator to replace Manifold.

smfgo prompts the user for inputs required for generating a basic and valid SMF
manifest file.

## Example Usage
```
$ smfgo myservice.xml
The name of the service bundle: mybundle
The name of the service: component/myservice
The version of the service manifest: 1
The human readable name of the service: My service.
Does this service support multiple instances? [y/N]: No
Full path to configuration file, blank if no configuration file: /etc/myconfig
Full command to start the service, use %{config/file} to reference configuration file, eg. '/usr/bin/myservice --start %{config/file}': /bin/myservice --start %{config/file}
Full command to stop the service, use %{config/file} to reference configuration file, eg. '/usr/bin/myservice --stop %{config/file}': :kill
Add environment variables for method execution? [y/N]: Yes
Environment variable name (leave blank to continue): HOME
Environment variable value: /home/service
Environment variable name (leave blank to continue):
Select service managment method:
  > child
User for execution methods: myuser
Group for execution methods: mygroup
Does this service depend on network being ready? [Y/n]: Yes
Does this service depend on local filesystems being ready? [Y/n]: Yes
Enable the service instance by default? [y/N]: No

$ cat myservice.xml
<?xml version='1.0'?>
<!DOCTYPE service_bundle SYSTEM '/usr/share/lib/xml/dtd/service_bundle.dtd.1'>
<service_bundle type="manifest" name="mybundle">
  <service name="component/myservice" type="service" version="1">
    <create_default_instance enabled="false"></create_default_instance>
    <single_instance></single_instance>
    <dependency name="network" grouping="require_all" restart_on="error" type="service">
      <service_fmri value="svc:/milestone/network:default"></service_fmri>
    </dependency>
    <dependency name="filesystem-local" grouping="require_all" restart_on="none" type="service">
      <service_fmri value="svc:/system/filesystem/local:default"></service_fmri>
    </dependency>
    <method_context>
      <method_credential user="myuser" group="mygroup"></method_credential>
      <method_environment>
        <envvar name="HOME" value="/home/service"></envvar>
      </method_environment>
    </method_context>
    <exec_method name="start" type="method" exec="/bin/myservice --start %{config/file}" timeout_seconds="60"></exec_method>
    <exec_method name="stop" type="method" exec=":kill" timeout_seconds="60"></exec_method>
    <property_group name="config" type="application">
      <propval name="file" type="astring" value="/etc/myconfig"></propval>
    </property_group>
    <property_group name="startd" type="framework">
      <propval name="duration" type="astring" value="child"></propval>
      <propval name="ignore_error" type="astring" value="core,signal"></propval>
    </property_group>
    <stability value="Evolving"></stability>
    <template>
      <common_name>
        <loctext xml:lang="C">My service.</loctext>
      </common_name>
    </template>
  </service>
</service_bundle>

$ svccfg validate myservice.xml
$ echo $?
0
```
