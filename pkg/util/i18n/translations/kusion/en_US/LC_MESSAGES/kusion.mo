��    =        S   �      8  0  9  ?   j  �   �  d  s  �  �  �  �  �   T  8  �  B     �   T  �   �  �   �  |   l  D  �  �  .  H   �  @     F   P  3   �  8   �  >     7   C     {  c   �  A   �  �   7     �  �   �  R   �  #   �  3   �  (   3  R   \  N   �  T   �  +   S  =     2   �  O   �  q   @  5   �  V   �  $   ?  $   d  $   �  ,   �  Z   �  &   6  1   ]     �     �     �     �  D   V     �     �     �  g   �  [   Y   ;   �   +  �   0  "  ?   N#  �   �#  d  W$  �  �%  �  �(  �   8+  8  �+  B   �,  �   8-  �   �-  �   �.  |   P/  D  �/  �  1  H   �2  @   �2  F   43  3   {3  8   �3  >   �3  7   '4     _4  c   u4  A   �4  �   5     �5  �   �5  R   l6  #   �6  3   �6  (   7  R   @7  N   �7  T   �7  +   78  =   c8  2   �8  O   �8  q   $9  5   �9  V   �9  $   #:  $   H:  $   m:  ,   �:  Z   �:  &   ;  1   A;     s;     �;     �;     �;  D   :<     <     �<     �<  g   �<  [   ==  ;   �=     0                     5   8      9   	                    +   2       .         /       &          1       -            =      4                         ;   !          *      
          3      #       <   "       (   )             '         %              ,   $      7              6   :              
		# Check configuration in main.k
		kusion check main.k

		# Check main.k with arguments
		kusion check main.k -D name=test -D age=18

		# Check main.k with arguments from settings.yaml
		kusion check main.k -Y settings.yaml

		# Check main.k with work directory
		kusion check main.k -w appops/demo/dev 
		# Delete the configuration of current stack
		kusion destroy 
		# Get name and description of internal templates
		kusion init templates

		# Get templates from specific templates repository
		kusion init templates https://github.com/<user>/<repo> --online=true 
		# Initialize a project from internal templates
		kusion init

		# Initialize a project from default online templates
		kusion init --online=true

		# Initialize a project from a specific online template
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a project from a specific local template
		kusion init /path/to/templates 
		# List all project and stack information in the current directory
		kusion ls

		# List all project and stack information in the specify directory
		kusion ls ./path/to/project_dir

		# List all project and stack information in the specify directory,
		# and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree

		# List all project and stack information in the specify directory,
		# and output in a JSON format
		kusion ls ./path/to/project_dir --format=json

		# List all project and stack information in the specify directory,
		# and output in a YAML format
		kusion ls ./path/to/project_dir --format=yaml

		# List all project and stack by level, and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree --level=1 
		# List all the KCL files that are dependent by the given focus paths
        kusion deps --focus path/to/focus1 --focus path/to/focus2

		# List all the projects that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2

		# List all the stacks that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --only stack

		# List all the projects that depend on the given focus paths, ignoring some paths from entrance files in each stack
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --ignore path/to/ignore 
		# Print Kusion environment information
		kusion env

		# Print Kusion environment information as JSON format
		kusion env --json 
		# The comparison object comes from the files
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# The comparison object comes from the stdin
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff - 
		Check if KCL configurations in current directory ok to compile. 
		Compare files differences and display the delta.
		Support input file types are: YAML (http://yaml.org/) and JSON (http://json.org/). 
		List all project and stack information in the current directory or the
		specify directory.
		The default output is in a human friendly format, and it also supports
		a variety of formatted structure output. 
		Show the KCL file dependency information in the current directory or the specified workdir.
    By default, it will list all the KCL files that are dependent on the given package path. 
		This command gets the descriptions and definitions of the templates which are used to initialize the project scaffolding. 
		This command initializes the scaffolding for a project, generating a project from an appointed template with correct structure.

		The scaffold templates can be retrieved from local or online. The built-in templates are used by default, self-defined templates are also supported by assigning the template repository path. 
    Env prints Kusion environment information.

    By default env prints information as a shell script (on Windows, a batch file). If one
    or more variable names is given as arguments, env prints the value of each named variable
    on its own line.

    The --json flag prints the environment in JSON format instead of as a shell script.

    For more about environment variables, see "kusion env -h". After creating/updating/deleting the requested object, watch for changes Automatically approve and perform the update after previewing it Automatically show all plan details, combined use with flag `--detail` Automatically show plan details after previewing it Automatically show plan details with interactive options Check if KCL configurations in current directory ok to compile Compare differences between input files <from> and <to> Compile KCL into YAML Custom params in JSON. If specified, it will be used as the template default value and skip prompts Diff mode. One of noemal and . The default is normal ignore-added Direct is set to "down", "only" means only the downstream project/stack list will be output. Valid values: project, stack. Defaults to project Disable dumping None values File paths to ignore when filtering the affected stacks/projects. Each path needs to be a valid relative path from the workdir. If not set, no paths will be ignored. Force generating the scaffolding files, even if it would change the existing files Ignore differences of target fields Ignore order changes in lists. The default is false Initialize the scaffolding for a project Initialize with specified project name. If not specified, a prompt will request it Initialize with specified template. If not specified, a prompt will request it Inspect direct of the dependency information. Valid values: up, down. Defaults to up List templates used to initialize a project Max display depth of the project and stack tree. One of 0,1,2 Omit the dyff summary header. The default is false Output format of the project information. valid values: json, yaml, tree, human Paths to focus on to inspect. It cannot be empty and each path needs to be a valid relative path from the workdir Preview a series of resource changes within the stack Preview the execution effect (always successful) without actually applying the changes Print Kusion environment information Print the environment in JSON format Show KCL file dependency information Skip prompts and proceed with default values Sort from and to by kubernetes resource order(non standard behavior). The default is false Specify the command line setting files Specify the configuration override path and value Specify the operator Specify the output file Specify the output format Specify the output format of templates. If specified, only support json for now; if not, template name and description is given Specify the output style. One of human and raw. The default is human Specify the override option Specify the top-level argument Specify the work directory Swap <from> and <to> for comparison. Note that it is invalid when <from> is stdin. The default is false Use templates from online repository to initialize project, or use locally cached templates no-style sets to RawOutput mode and disables all of styling Project-Id-Version: 
Report-Msgid-Bugs-To: EMAIL
PO-Revision-Date: 2023-07-11 21:46+0800
Last-Translator: Automatically generated
Language-Team: none
Language: en_US
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit
Plural-Forms: nplurals=2; plural=(n != 1);
 
		# Check configuration in main.k
		kusion check main.k

		# Check main.k with arguments
		kusion check main.k -D name=test -D age=18

		# Check main.k with arguments from settings.yaml
		kusion check main.k -Y settings.yaml

		# Check main.k with work directory
		kusion check main.k -w appops/demo/dev 
		# Delete the configuration of current stack
		kusion destroy 
		# Get name and description of internal templates
		kusion init templates

		# Get templates from specific templates repository
		kusion init templates https://github.com/<user>/<repo> --online=true 
		# Initialize a project from internal templates
		kusion init

		# Initialize a project from default online templates
		kusion init --online=true

		# Initialize a project from a specific online template
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a project from a specific local template
		kusion init /path/to/templates 
		# List all project and stack information in the current directory
		kusion ls

		# List all project and stack information in the specify directory
		kusion ls ./path/to/project_dir

		# List all project and stack information in the specify directory,
		# and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree

		# List all project and stack information in the specify directory,
		# and output in a JSON format
		kusion ls ./path/to/project_dir --format=json

		# List all project and stack information in the specify directory,
		# and output in a YAML format
		kusion ls ./path/to/project_dir --format=yaml

		# List all project and stack by level, and output in a Tree format
		kusion ls ./path/to/project_dir --format=tree --level=1 
		# List all the KCL files that are dependent by the given focus paths
        kusion deps --focus path/to/focus1 --focus path/to/focus2

		# List all the projects that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2

		# List all the stacks that depend on the given focus paths
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --only stack

		# List all the projects that depend on the given focus paths, ignoring some paths from entrance files in each stack
		kusion deps --direct down --focus path/to/focus1 --focus path/to/focus2 --ignore path/to/ignore 
		# Print Kusion environment information
		kusion env

		# Print Kusion environment information as JSON format
		kusion env --json 
		# The comparison object comes from the files
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# The comparison object comes from the stdin
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff - 
		Check if KCL configurations in current directory ok to compile. 
		Compare files differences and display the delta.
		Support input file types are: YAML (http://yaml.org/) and JSON (http://json.org/). 
		List all project and stack information in the current directory or the
		specify directory.
		The default output is in a human friendly format, and it also supports
		a variety of formatted structure output. 
		Show the KCL file dependency information in the current directory or the specified workdir.
    By default, it will list all the KCL files that are dependent on the given package path. 
		This command gets the descriptions and definitions of the templates which are used to initialize the project scaffolding. 
		This command initializes the scaffolding for a project, generating a project from an appointed template with correct structure.

		The scaffold templates can be retrieved from local or online. The built-in templates are used by default, self-defined templates are also supported by assigning the template repository path. 
    Env prints Kusion environment information.

    By default env prints information as a shell script (on Windows, a batch file). If one
    or more variable names is given as arguments, env prints the value of each named variable
    on its own line.

    The --json flag prints the environment in JSON format instead of as a shell script.

    For more about environment variables, see "kusion env -h". After creating/updating/deleting the requested object, watch for changes Automatically approve and perform the update after previewing it Automatically show all plan details, combined use with flag `--detail` Automatically show plan details after previewing it Automatically show plan details with interactive options Check if KCL configurations in current directory ok to compile Compare differences between input files <from> and <to> Compile KCL into YAML Custom params in JSON. If specified, it will be used as the template default value and skip prompts Diff mode. One of noemal and . The default is normal ignore-added Direct is set to "down", "only" means only the downstream project/stack list will be output. Valid values: project, stack. Defaults to project Disable dumping None values File paths to ignore when filtering the affected stacks/projects. Each path needs to be a valid relative path from the workdir. If not set, no paths will be ignored. Force generating the scaffolding files, even if it would change the existing files Ignore differences of target fields Ignore order changes in lists. The default is false Initialize the scaffolding for a project Initialize with specified project name. If not specified, a prompt will request it Initialize with specified template. If not specified, a prompt will request it Inspect direct of the dependency information. Valid values: up, down. Defaults to up List templates used to initialize a project Max display depth of the project and stack tree. One of 0,1,2 Omit the dyff summary header. The default is false Output format of the project information. valid values: json, yaml, tree, human Paths to focus on to inspect. It cannot be empty and each path needs to be a valid relative path from the workdir Preview a series of resource changes within the stack Preview the execution effect (always successful) without actually applying the changes Print Kusion environment information Print the environment in JSON format Show KCL file dependency information Skip prompts and proceed with default values Sort from and to by kubernetes resource order(non standard behavior). The default is false Specify the command line setting files Specify the configuration override path and value Specify the operator Specify the output file Specify the output format Specify the output format of templates. If specified, only support json for now; if not, template name and description is given Specify the output style. One of human and raw. The default is human Specify the override option Specify the top-level argument Specify the work directory Swap <from> and <to> for comparison. Note that it is invalid when <from> is stdin. The default is false Use templates from online repository to initialize project, or use locally cached templates no-style sets to RawOutput mode and disables all of styling 