## kusion env

Print Kusion environment information

### Synopsis

Env prints Kusion environment information.

 By default env prints information as a shell script (on Windows, a batch file). If one or more variable names is given as arguments, env prints the value of each named variable on its own line.

 The --json flag prints the environment in JSON format instead of as a shell script.

 For more about environment variables, see "kusion env -h".

```
kusion env [flags]
```

### Examples

```
  # Print Kusion environment information
  kusion env
  
  # Print Kusion environment information as JSON format
  kusion env --json
```

### Options

```
  -h, --help   help for env
      --json   以 JSON 格式打印环境信息
```

### SEE ALSO

* [kusion](kusion.md)	 - kusion 通过代码管理 Kubernetes 集群

###### Auto generated by spf13/cobra on 13-Jul-2023