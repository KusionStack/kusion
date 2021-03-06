## kusion preview

Preview a series of resource changes within the stack.

### Synopsis

Preview a series of resource changes within the stack.

Create or update or delete resources according to the KCL files within a stack. By default, Kusion will generate an execution plan and present it for your approval before taking any action.

```
kusion preview [flags]
```

### Examples

```
  # Preview with specifying work directory
  kusion preview -w /path/to/workdir
  
  # Preview with specifying arguments
  kusion preview -D name=test -D age=18
  
  # Preview with specifying setting file
  kusion preview -Y settings.yaml
```

### Options

```
  -D, --argument strings    Specify the arguments to preview KCL
  -d, --detail              Automatically show plan details after previewing it
  -h, --help                help for preview
      --no-style            no-style sets to RawOutput mode and disables all of styling
  -O, --overrides strings   Specify the configuration override path and value
  -Y, --setting strings     Specify the command line setting files
  -w, --workdir string      Specify the work directory
  -y, --yes                 Show preview only, no details
```

### SEE ALSO

* [kusion](kusion.md)	 - kusion manages the Kubernetes cluster by code

###### Auto generated by spf13/cobra on 29-Apr-2022
