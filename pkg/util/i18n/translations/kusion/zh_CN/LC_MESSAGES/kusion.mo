Þ    A      $  Y   ,        y    0    Y  <  È   
  d  _  ù  Ä  ã  ¾     ¢  .   &  8  U  Y    B   è     +  Ò   ´  ý     ¢     |   (  D  ¥  B   ê  H   -  D   v  @   »  F   ü  3   C  >   w  7   ¶     î  c     4   h       "   ¹  R   Ü  #   /  3   S  (     R   °  N     T   R  ,   §  &   Ô  +   û  =   '  2   e  O     5   è  V      $   u         ,   ¸   Z   å   &   @!  1   g!     !     ®!     Æ!  D   F"     "     §"     Æ"  g   á"  [   I#  +   ¥#  *   Ñ#  ;   ü#  Z  8$  C  %  -  ×&  S  (  ¼   Y*  ^  +  ¾  u,  T  4/  o   0  *   ù0    $1  H  ?2  ;   3  z   Ä3     ?4  á   Ú4     ¼5  Q   L6    6  1   ¥7  8   ×7  l   8  $   }8  '   ¢8  (   Ê8  5   ó8  2   )9     \9  h   s9  1   Ü9     :     $:  9   C:     }:  -   :  $   Ç:  N   ì:  E   ;;  F   ;  &   È;  %   ï;  '   <  B   =<  '   <  F   ¨<  (   ï<  E   =     ^=     y=  /   =  E   Ç=     >     )>     H>     X>  w   k>  ;   ã>     ?     2?     E?  [   X?  D   ´?  -   ù?  L   '@  6   t@                   (       3       7   <   :         =          "      6                       @   ;       )   +   .           -   4   ?       A             !                 ,   2   	                8       >       1          *      /                 0      
      %       #              5   9   '         &             $        
		# Apply with specifying work directory
		kusion apply -w /path/to/workdir
	
		# Apply with specifying arguments
		kusion apply -D name=test -D age=18
	
		# Apply with specifying setting file
		kusion apply -Y settings.yaml
	
		# Skip interactive approval of plan details before applying
		kusion apply --yes
		
		# Apply without output style and color
		kusion apply -s=true 
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
		kusion compile main.k -o output.yaml
		
		# Complie without output style and color
		kusion compile -s=true 
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
		# Preview with specifying work directory
		kusion preview -w /path/to/workdir
	
		# Preview with specifying arguments
		kusion preview -D name=test -D age=18
	
		# Preview with specifying setting file
		kusion preview -Y settings.yaml
	
		# Preview with ignored fields
		kusion preview --ignore-fields="metadata.generation,metadata.managedFields
		
		# Preview with json format result
		kusion preview -o json
		
		# Preview without output style and color
		kusion preview -s=true 
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
		List all project and stack information in the current directory or the
		specify directory.
		The default output is in a human friendly format, and it also supports
		a variety of formatted structure output. 
		Preview a series of resource changes within the stack.
	
		Create or update or delete resources according to the KCL files within a stack. By default,
		Kusion will generate an execution plan and present it for your approval before taking any action. 
		The Kusion command line tool is the entry point for users to use KusionStack. 
		Users can use the Kusion command line tool to complete all operation tasks.
		 
		This command gets the descriptions and definitions of the templates which are used to initialize the project scaffolding. 
		This command initializes the scaffolding for a project, generating a project from an appointed template with correct structure.

		The scaffold templates can be retrieved from local or online. The built-in templates are used by default, self-defined templates are also supported by assigning the template repository path. 
    Print the kusion version information for the current context. After creating/updating/deleting the requested object, watch for changes Apply the operation intents of various resources to multiple runtime Automatically approve and perform the update after previewing it Automatically show all plan details, combined use with flag `--detail` Automatically show plan details after previewing it Check if KCL configurations in current directory ok to compile Compare differences between input files <from> and <to> Compile KCL into YAML Custom params in JSON. If specified, it will be used as the template default value and skip prompts Delete the resources of a specified stack in runtime Disable dumping None values Disable the output style and color Force generating the scaffolding files, even if it would change the existing files Ignore differences of target fields Ignore order changes in lists. The default is false Initialize the scaffolding for a project Initialize with specified project name. If not specified, a prompt will request it Initialize with specified template. If not specified, a prompt will request it Inspect direct of the dependency information. Valid values: up, down. Defaults to up Kusion is the entry point to use KusionStack List all project and stack information List templates used to initialize a project Max display depth of the project and stack tree. One of 0,1,2 Omit the dyff summary header. The default is false Output format of the project information. valid values: json, yaml, tree, human Preview a series of resource changes within the stack Preview the execution effect (always successful) without actually applying the changes Print Kusion environment information Print the kusion version info Skip prompts and proceed with default values Sort from and to by kubernetes resource order(non standard behavior). The default is false Specify the command line setting files Specify the configuration override path and value Specify the operator Specify the output file Specify the output format of templates. If specified, only support json for now; if not, template name and description is given Specify the output style. One of human and raw. The default is human Specify the override option Specify the top-level argument Specify the work directory Swap <from> and <to> for comparison. Note that it is invalid when <from> is stdin. The default is false Use templates from online repository to initialize project, or use locally cached templates backend-config config state storage backend backend-type specify state storage backend no-style sets to RawOutput mode and disables all of styling Project-Id-Version: gettext-go-examples-hello
Report-Msgid-Bugs-To: EMAIL
PO-Revision-Date: 2023-07-13 21:55+0800
Last-Translator: howieyuen <howieyuen@outlook.com>
Language-Team: 
Language: zh
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: 8bit
Plural-Forms: nplurals=1; plural=0;
X-Generator: Poedit 3.3.2
 
		# æå®å·¥ä½ç®å½ apply
		kusion apply -w /path/to/workdir

		# æå®åæ° apply
		kusion apply -D name=test -D age=18

		# æå®éç½®æä»¶ apply
		kusion apply -Y settings.yaml

		# apply åè·³è¿è®¡åçäº¤äºå¼å®¡æ¹
		kusion apply âyes

		# apply è¾åºç»æä¸å¸¦æ ·å¼åé¢è²
		kusion apply -s=true 
		# éªè¯ main.k ä¸­çéç½®
		kusion check main.k

		# ä½¿ç¨åæ°éªè¯ main.k
		kusion check main.k -D name=test -D age=18

		# ä½¿ç¨æ¥èª settings.yaml çåæ°éªè¯ main.k
		kusion check main.k -Y settings.yaml

		# ä½¿ç¨å·¥ä½ç®å½éªè¯ main.k
		kusion check main.k -w appops/demo/dev 
		# å° main.k ä¸­çéç½®ç¼è¯æ YAML æ ¼å¼
		kusion compile main.k

		# ä½¿ç¨åæ°ç¼è¯ main.k
		kusion compile main.k -D name=test -D age=18

		# ä½¿ç¨æ¥èª settings.yaml çåæ°ç¼è¯ main.k
		kusion compile main.k -Y settings.yaml

		# ä½¿ç¨å·¥ä½ç®å½ç¼è¯ main.k
		kusion compile main.k -w appops/demo/dev

		# Override å½å Stack ä¸­ç KCL éç½®
		kusion compile -O __main__:appConfiguration.image=nginx:latest -a

		# ç¼è¯ main.k å¹¶å°ç»æåå¥ output.yaml
		kusion compile main.k -o output.yaml
		
		# ç¼è¯è¾åºä¸å¸¦æ ·å¼åé¢è²
		kusion compile -s=true 
		# è·ååç½®æ¨¡æ¿çåç§°åæè¿°
		kusion init templates

		# è·åæå®çå¨çº¿ä»åºä¸­çæ¨¡æ¿åè¡¨
		kusion init templates https://github.com/<user>/<repo> âonline=true 
		# ä»æ¬å°åç½®æ¨¡æ¿ä¸­åå§åä¸ä¸ª project
		kusion init

		# ä»é»è®¤çå¨çº¿æ¨¡æ¿ä¸­åå§åä¸ä¸ª project
		kusion init --online=true

		# ä»æå®çå¨çº¿æ¨¡æ¿åå§åä¸ä¸ª project
		kusion init https://github.com/<user>/<repo> --online=true

		# ä»æå®çæ¬å°æ¨¡æ¿åå§åä¸ä¸ª project
		kusion init /path/to/templates 
		# ååºå½åç®å½ä¸çææ project å stack ä¿¡æ¯
		kusion ls

		# ååºæå®ç®å½ä¸­çææ project å stack ä¿¡æ¯
		kusion ls ./path/to/project_dir

		# ååºæå®ç®å½ä¸çææ project å stack ä¿¡æ¯ï¼å¹¶ä»¥ Tree æ ¼å¼è¾åº
		kusion ls ./path/to/project_dir --format=tree

		# ååºæå®ç®å½ä¸çææ project å stack ä¿¡æ¯ï¼å¹¶ä»¥ JSON æ ¼å¼è¾åº
		kusion ls ./path/to/project_dir --format=json

		# ååºæå®ç®å½ä¸çææ project å stack ä¿¡æ¯ï¼å¹¶ä»¥ YAML æ ¼å¼è¾åº
		kusion ls ./path/to/project_dir --format=yaml

		# éçº§ååºææ project å stackï¼å¹¶ä»¥ Tree æ ¼å¼è¾åº
		kusion ls ./path/to/project_dir --format=tree --level=1 
		# ä½¿ç¨æå®å·¥ä½ç®å½è¿è¡é¢è§
		kusion preview -w /path/to/workdir

		# ä½¿ç¨æå®åæ°é¢è§
		kusion preview -D name=test -D age=18

		# æå®è®¾ç½®æä»¶é¢è§
		kusion preview -Y settings.yaml

		# é¢è§ç»æè®¾ç½®ä¸º JSON æ ¼å¼
		kusion preview -o json

		# é¢è§ç»æä¸å¸¦æ ·å¼åé¢è²
		kusion preview -s=true 
		# æå° Kusion ç¯å¢ä¿¡æ¯
		kusion env

		# ä»¥ JSON æ ¼å¼æå° Kusion ç¯å¢ä¿¡æ¯
		kusion env --json 
		# æå° Kusion çæ¬
		kusion version 
		# æ¯è¾å¯¹è±¡æ¥èªæä»¶
		kusion diff pod-1.yaml pod-2.yaml
		kusion diff pod-1.yaml pod-2.yaml --swap=true

		# æ¯è¾å¯¹è±¡æ¥èªæ åè¾å¥
		cat pod-1.yaml > pod-full.yaml
		echo '---' >> pod-full.yaml
		cat pod-2.yaml >> pod-full.yaml
		cat pod-full.yaml | kusion diff - 
		å°ä¸ç³»åèµæºåæ´ apply å°å½åæ ä¸­ã

		å¨å½å Stack ä¸­æ ¹æ® KCL æä»¶åå»ºãæ´æ°ãå é¤èµæºã
		é»è®¤æåµä¸ï¼å¨éåä»»ä½è¡å¨ä¹åï¼Kusion ä¼çæä¸ä¸ªæ§è¡è®¡åï¼å¹¶æäº¤ç»ä½ æ¹åã

		ä½ å¯ä»¥æ£æ¥è®¡åè¯¦ç»ä¿¡æ¯ï¼ç¶åå³å®æ¯å¦åºéåæä¸­æ­¢è¿äºæä½ã 
		æ£æ¥å½åç®å½ä¸­ç KCL éç½®æ¯å¦å¯ä»¥ç¼è¯ã 
		æ¯è¾æä»¶å·®å¼å¹¶æ¾ç¤ºå¢éã
		æ¯æçæä»¶ç±»åæ: YAML (http://yaml.org/) å JSON (http://json.org/)ã 
		ååºå½åç®å½ææå®ç®å½ä¸­çææ project å stack ä¿¡æ¯ã
		é»è®¤è¾åºæ¯äººæ§åçæ ¼å¼ï¼ä¹æ¯æå¤ç§æ ¼å¼çç»æè¾åºã 
		é¢è§ Stack ä¸­çä¸ç³»åèµæºæ´æ¹ã

		æ ¹æ® Stack åç KCL æä»¶åå»ºææ´æ°æå é¤èµæºã
		é»è®¤æåµä¸ï¼Kusion ä¼çæä¸ä¸ªæ§è¡è®¡åå¹¶å¨éåä»»ä½è¡å¨ä¹åå°å¶æäº¤ç»æ¨æ¹åã 
		Kusion å½ä»¤è¡å·¥å·æ¯ç¨æ·ä½¿ç¨ KusionStack çå¥å£ã 
		ç¨æ·å¯ä»¥ä½¿ç¨ Kusion å½ä»¤è¡å·¥å·å®æææçè¿ç»´æä½ã
		 
		è¯¥æä»¤è·åç¨äºåå§å project èææ¶çæ¨¡æ¿çæè¿°åå®ä¹ã 
		è¯¥æä»¤ç¨äºåå§åä¸ä¸ª project èææ¶ï¼ä»éå®çæ¨¡æ¿çæå·å¤æ­£ç¡®ç»æç project.

		èææ¶æ¨¡æ¿å¯ä»¥ä»æ¬å°æèçº¿ä¸è·åãé»è®¤ä½¿ç¨åç½®æ¨¡æ¿ï¼è¥éè¦ä½¿ç¨èªå®ä¹çæ¨¡æ¿ï¼éè¦æå®æ¨¡æ¿ä»åºçå°åã 
    æå°å½åä¼è¯ Kusion ççæ¬ä¿¡æ¯ã å¨åå»º/æ´æ°/å é¤è¯·æ±çå¯¹è±¡åï¼è§æµåæ´ å°åç±»èµæºçè¿ç»´æå¾ apply è³å¤ç§è¿è¡æ¶é (KubernetesãAWS/é¿éäºä»¥åèªå»ºç³»ç»ç­) é¢è§åèªå¨å®¡æ¹å¹¶åºç¨æ´æ° é¢è§åèªå¨å±ç¤ºæ§è¡è®¡åç»è é¢è§åèªå¨å±ç¤º apply è®¡åç»è æ£æ¥å½åç®å½ä¸­ç KCL éç½®æ¯å¦å¯ä»¥ç¼è¯ æ¯è¾è¾å¥æä»¶ <from> å <to> ä¹é´çå·®å¼ å° KCL ç¼è¯æ YAML JSON æ ¼å¼çèªå®ä¹åæ°ãè¥æå®ï¼é£ä¹ä¼ç¨å¶ä½ä¸ºé»è®¤çæ¨¡æ¿åæ°ï¼å¹¶è·³è¿æç¤º å é¤ä¸ä¸ªç¹å® stack å¨è¿è¡æ¶ä¸­çèµæº ç¦ç¨è½¬å¨ None å¼ ç¦ç¨è¾åºçæ ·å¼ä¸é¢è² å¼ºå¶çæèææ¶æä»¶ï¼å³ä½¿ä¼æ´æ¹ç°ææä»¶ å¿½ç¥ç®æ å­æ®µçå·®å¼ å¿½ç¥åè¡¨ä¸­çé¡ºåºååï¼é»è®¤ä¸ºå¦ åå§åä¸ä¸ª project çèææ¶ ç¨æå®ç project åç§°åå§åãè¥ä¸æå®ï¼å°ææç¤ºè¯·æ±è·å ç¨æå®çæ¨¡æ¿åå§åãè¥ä¸æå®ï¼å°ææç¤ºè¯·æ±è·å project ä¿¡æ¯çè¾åºæ ¼å¼ãå¯éå¼ï¼jsonãyamlãtreeãhuman Kusion æ¯ä½¿ç¨ KusionStack çå¥å£ ååºææ project å stack ä¿¡æ¯ ååºç¨äºåå§å project çæ¨¡æ¿ project å stack æ çæå¤§æ¾ç¤ºæ·±åº¦ãå¯éå¼ï¼0ã1ã2 çç¥ dyff æè¦æ é¢ï¼é»è®¤ä¸ºå¦ project ä¿¡æ¯çè¾åºæ ¼å¼ãå¯éå¼ï¼jsonãyamlãtreeãhuman é¢è§ Stack ä¸­çä¸ç³»åèµæºæ´æ¹ dry-run é¢è§æ§è¡ææï¼æ»æ¯æåï¼èä¸å®é apply æ´æ¹ æå° Kusion ç¯å¢ä¿¡æ¯ æå° Kusion ççæ¬ä¿¡æ¯ è·³è¿æç¤ºï¼ä½¿ç¨é»è®¤å¼åå§å project æ kubernetes èµæºé¡ºåºæåºï¼éæ åè¡ä¸ºï¼ãé»è®¤ä¸ºå¦ æå®å½ä»¤è¡éç½®æä»¶ æå®éç½®è¦çè·¯å¾åå¼ æå®æä½äºº æå®è¾åºæä»¶ æå®æ¨¡æ¿çè¾åºæ ¼å¼ãè¥æå®ï¼ç°å¨ä»æ¯æ jsonï¼è¥ä¸æå®ï¼é£ä¹å°æå°æ¨¡æ¿çåç§°åæè¿° æå®è¾åºé£æ ¼ï¼human å raw ä¹ä¸ï¼é»è®¤å¼ human æå®è¦çéé¡¹ æå®é¡¶çº§åæ° æå®å·¥ä½ç®å½ äº¤æ¢ <from> å <to> è¿è¡æ¯è¾ãæ³¨æ <from> ä¸ºæ åè¾å¥æ¶æ æãé»è®¤ä¸ºå¦ ä½¿ç¨çº¿ä¸æ¨¡æ¿åå§å projectï¼å¦åä½¿ç¨æ¬å°ç¼å­æ¨¡æ¿ backend-config éç½®ç¶æå­å¨åç«¯ä¿¡æ¯ backend-type æå®ç¶æå­å¨åç«¯ç±»åï¼æ¯æ localãdbãoss å s3 no-style è®¾ç½®åå§è¾åºæ¨¡å¼å¹¶ç¦ç¨æææ ·å¼ 