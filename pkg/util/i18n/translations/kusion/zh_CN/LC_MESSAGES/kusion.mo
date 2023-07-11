��    -      �  =   �      �  3  �  0      F  �  T  �  �	  �   �  .   Y  8  �  W  �  B     �   \  �   �  �  �  �   T  <   '  @   d  3   �  >   �  7        P  >   f     �  E   �  3     >   ;  &   z  2   �  5   �  $   
     /  ,   M  Z   z  &   �  1   �     .     C  D   [     �     �     �  g   �  <   ^  @   �  ;   �  X    �   q  -  q    �  U  �  �  !  o   �#  *   $    J$  F  e%  ;   �&  z   �&  �   c'  P  2(  �   �)  <   *  $   W*  (   |*  5   �*  2   �*     +  6   %+     \+  3   r+  -   �+  A   �+  !   ,  '   8,  '   `,     �,     �,  $   �,  E   �,     ,-     H-     g-     w-  ;   �-     �-     �-     �-  [   �-  3   [.  6   �.  6   �.        +   -                        *         #         	   (                               &                        %   $   
       '             ,                     !         )      "                              
		# Apply with specifying work directory
		kusion apply -w /path/to/workdir

		# Apply with specifying arguments
		kusion apply -D name=test -D age=18

		# Apply with specifying setting file
		kusion apply -Y settings.yaml

		# Skip interactive approval of plan details before applying
		kusion apply --yes 
		# Check configuration in main.k
		kusion check main.k

		# Check main.k with arguments
		kusion check main.k -D name=test -D age=18

		# Check main.k with arguments from settings.yaml
		kusion check main.k -Y settings.yaml

		# Check main.k with work directory
		kusion check main.k -w appops/demo/dev 
		# Compile configuration in main.k into YAML format
		kusion compile main.k

		# Compile main.k with arguments
		kusion compile main.k -D name=test -D age=18

		# Compile main.k with arguments from settings.yaml
		kusion compile main.k -Y settings.yaml

		# Compile main.k with work directory
		kusion compile main.k -w appops/demo/dev

		# Compile with override
		kusion compile -O __main__:appConfiguration.image=nginx:latest -a

		# Compile main.k and write result into output.yaml
		kusion compile main.k -o output.yaml 
		# Initialize a new KCL project from internal templates
		kusion init

		# Initialize a new KCL project from external default templates location
		kusion init --online=true

		# Initialize a new KCL project from specific templates location
		kusion init https://github.com/<user>/<repo> --online=true

		# Initialize a new KCL project from local directory
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
		# Print Kusion environment information
		kusion env

		# Print Kusion environment information as JSON format
		kusion env --json 
		# Print the kusion version
		kusion version 
		# The comparison object comes from the files
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# The comparison object comes from the stdin
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff - 
		Apply a series of resource changes within the stack.

		Create or update or delete resources according to the KCL files within a stack.
		By default, Kusion will generate an execution plan and present it for your approval before taking any action.

		You can check the plan details and then decide if the actions should be taken or aborted. 
		Check if KCL configurations in current directory ok to compile. 
		Compare files differences and display the delta.
		Support input file types are: YAML (http://yaml.org/) and JSON (http://json.org/). 
		Compile one or more KCL files.

		The KCL filename must be specified.
		You can specify a list of arguments to replace the placeholders defined in KCL,
		and output the compiled results to a file when using --output flag. 
		Delete resources by resource spec.

		Only KCL files are accepted. Only one type of arguments may be specified: filenames,
		resources and names, or resources and label selector.

		Note that the destroy command does NOT do resource version checks, so if someone submits an
		update to a resource right when you submit a destroy, their update will be lost along with the
		rest of the resource. 
		List all project and stack information in the current directory or the
		specify directory.
		The default output is in a human friendly format, and it also supports
		a variety of formatted structure output. Apply a configuration stack to resource(s) by work directory Automatically approve and perform the update after previewing it Automatically show plan details after previewing it Check if KCL configurations in current directory ok to compile Compare differences between input files <from> and <to> Compile KCL into YAML Destroy a configuration stack to resource(s) by work directory Disable dumping None values Forces content to be generated even if it would change existing files Ignore order changes in lists. The default is false Initialize KCL file structure and base codes for a new project List all project and stack information Omit the dyff summary header. The default is false Preview a series of resource changes within the stack Print Kusion environment information Print the kusion version info Skip prompts and proceed with default values Sort from and to by kubernetes resource order(non standard behavior). The default is false Specify the command line setting files Specify the configuration override path and value Specify the operator Specify the output file Specify the output style. One of human and raw. The default is human Specify the override option Specify the top-level argument Specify the work directory Swap <from> and <to> for comparison. Note that it is invalid when <from> is stdin. The default is false The project name; if not specified, a prompt will request it Use locally cached templates without making any network requests no-style sets to RawOutput mode and disables all of styling Project-Id-Version: gettext-go-examples-hello
Report-Msgid-Bugs-To: EMAIL
PO-Revision-Date: 2022-11-22 16:48+0800
Last-Translator: howieyuen <howieyuen@outlook.com>
Language-Team: 
Language: zh
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit
Plural-Forms: nplurals=1; plural=0;
X-Generator: Poedit 3.2
 
		# 指定工作目录应用
		kusion apply -w /path/to/workdir

		# 指定参数应用
		kusion apply -D name=test -D age=18

		# 指定配置文件应用
		kusion apply -Y settings.yaml

		# 应用前跳过计划的交互式审批
		kusion apply —yes 
		# 验证 main.k 中的配置
		kusion check main.k

		# 使用参数验证 main.k
		kusion check main.k -D name=test -D age=18

		# 使用来自 settings.yaml 的参数验证 main.k
		kusion check main.k -Y settings.yaml

		# 使用工作目录验证 main.k
		kusion check main.k -w appops/demo/dev 
		# 将 main.k 中的配置编译成 YAML 格式
		kusion compile main.k

		# 使用参数编译 main.k
		kusion compile main.k -D name=test -D age=18

		# 使用来自 settings.yaml 的参数编译 main.k
		kusion compile main.k -Y settings.yaml

		# 使用工作目录编译 main.k
		kusion compile main.k -w appops/demo/dev

		# Override 当前 Stack 中的 KCL 配置
		kusion compile -O __main__:appConfiguration.image=nginx:latest -a

		# 编译 main.k 并将结果写入 output.yaml
		kusion compile main.k -o output.yaml 
		# 从内部模板初始化新的 KCL 项目
		kusion init

		# 从外部默认模板仓库初始化新的 KCL 项目
		kusion init --online=true

		# 从指定模板仓库初始化新的 KCL 项目
		kusion init https://github.com/<user>/<repo> --online=true

		# 从本地目录初始化新的 KCL 项目
		kusion init /path/to/templates 
		# 列出当前目录下的所有项目和堆栈信息
		kusion ls

		# 列出指定目录中的所有项目和堆栈信息
		kusion ls ./path/to/project_dir

		# 列出指定目录下的所有项目和堆栈信息，并以 Tree 格式输出
		kusion ls ./path/to/project_dir --format=tree

		# 列出指定目录下的所有项目和堆栈信息，并以 JSON 格式输出
		kusion ls ./path/to/project_dir --format=json

		# 列出指定目录下的所有项目和堆栈信息，并以 YAML 格式输出
		kusion ls ./path/to/project_dir --format=yaml

		# 逐级列出所有项目和堆栈，并以 Tree 格式输出
		kusion ls ./path/to/project_dir --format=tree --level=1 
		# 打印 Kusion 环境信息
		kusion env

		# 以 JSON 格式打印 Kusion 环境信息
		kusion env --json 
		# 打印 kusion 版本
		kusion version 
		# 比较对象来自文件
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# 比较对象来自标准输入
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff - 
		将一系列资源变更应用到当前栈中。

		在当前 Stack 中根据 KCL 文件创建、更新、删除资源。
		默认情况下，在采取任何行动之前，kusion会生成一个执行计划，并提交给你批准。

		你可以检查计划详细信息，然后决定是否应采取或中止这些操作。 
		检查当前目录中的 KCL 配置是否可以编译。 
		比较文件差异并显示增量。
		支持的文件类型有: YAML (http://yaml.org/) 和 JSON (http://json.org/)。 
		编译一个或多个 KCL 文件。

		必须指定 KCL 文件名。
		你可以指定参数列表来替换 KCL 中定义的占位符，
		并在使用 --output 标志时将编译结果输出到文件。 
		通过资源规约删除资源。

		只接受 KCL 文件。只能指定一种类型的参数：文件名、资源、名称、资源或标签选择器。

		请注意，destroy 命令不会进行资源版本检查，
		因此如果有人在你提交销毁时提交了对资源的更新，
		他们的更新将与资源一起丢失。 
		列出当前目录或指定目录中的所有项目和堆栈信息。
		默认输出是人性化的格式，也支持多种格式的结构输出。 通过工作目录将配置栈应用到一个或多个资源 预览后自动审批并应用更新 预览后自动展示 apply 计划细节 检查当前目录中的 KCL 配置是否可以编译 比较输入文件 <from> 和 <to> 之间的差异 将 KCL 编译成 YAML 通过工作目录销毁已应用到资源的配置栈 禁用转储 None 值 强制生成内容，即使它会更改现有文件 忽略列表中的顺序变化，默认为否 初始化一个新项目，包含 KCL 文件结构和基础代码 列出所有项目和堆栈信息 省略 dyff 摘要标题，默认为否 预览堆栈中的一系列资源更改 打印 Kusion 环境信息 打印 kusion 的版本信息 跳过提示并继续使用默认值 按 kubernetes 资源顺序排序（非标准行为）。默认为否 指定命令行配置文件 指定配置覆盖路径和值 指定操作人 指定输出文件 指定输出风格，human 和 raw 之一，默认值 human 指定覆盖选项 指定顶级参数 指定工作目录 交换 <from> 和 <to> 进行比较。注意 <from> 为标准输入时无效。默认为否 项目名称；若未指定，提示将要求该值 使用本地缓存模板而不发出任何网络请求 no-style 设置原始输出模式并禁用所有样式 