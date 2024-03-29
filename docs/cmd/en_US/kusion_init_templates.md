## kusion init templates

List templates used to initialize a project

### Synopsis

This command gets the descriptions and definitions of the templates which are used to initialize the project scaffolding.

```
kusion init templates
```

### Examples

```
  # Get name and description of internal templates
  kusion init templates
  
  # Get templates from specific templates repository
  kusion init templates https://github.com/<user>/<repo> --online=true
```

### Options

```
  -h, --help            help for templates
  -o, --output string   Specify the output format of templates. If specified, only support json for now; if not, template name and description is given
```

### Options inherited from parent commands

```
      --online                  Use templates from online repository to initialize project, or use locally cached templates
      --profile string          Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string   Name of the file to write the profile to (default "profile.pprof")
```

### SEE ALSO

* [kusion init](kusion_init.md)	 - Initialize the scaffolding for a project

###### Auto generated by spf13/cobra on 4-Jan-2024
