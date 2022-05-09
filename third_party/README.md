# `/third_party`

## Grumpy
Grumpy consists of a trans-compiler, and a list of runtime apis working like cpython. kusion will use the runtime part which is verified in Google's prod. However, kusion doesn't follow the trans-compiler approach, works in a standard compiler-vm way, we have to do some code change in grumpy codebase to supply the demand. Code changes may include coding convenience, fitting to python3 behaviors, implementing missing functions, bug fixes, etc. Kusion extension code will be placed in *_kusion.go files.  
