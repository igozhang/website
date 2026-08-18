# ai02_lm-studio

## env

lm-studio 0.3.10.0  
4070 laptop

## issues

1. 下载大模型文件GGUF  
   下载完后需要载指定目录递增两层英文文件名才能被准确识别，比如我指定的模型文件夹为D:\models，下载了DeepSeek-R1-Distill-Qwen-1.5B-Q2\_K.gguf文件的详细位置放在"D:\models\L1\L2\DeepSeek-R1-Distill-Qwen-1.5B-Q2\_K.gguf"才被准确识别；
2. 国内默认无法用discover搜索到模型，可以通过换源解决  
   思路： 替换huggingface.co 为 hf-mirror.com

2.1 打开文件位置  
在lm-studio上右键选择打开文件位置，两次，第一次是到快捷方式位置，第二次才到以下位置  
C:\Users\igozh\AppData\Local\Programs\LM Studio

2.2 打开powershell  
进入目录

```
> cd 'C:\Users\igozh\AppData\Local\Programs\LM Studio\'
> findstr /s /i /m /c:"huggingface.co" *.*
resources\app\.webpack\lms.exe
resources\app\.webpack\main\index.js
resources\app\.webpack\renderer\main_window.js
resources\app\.webpack\renderer\main_window.js.map
```

2.3 替换源  
在上一步找到的文件中替换huggingface.co 为 hf-mirror.com  
里面有个lms.exe忽略，其他文件替换

2.4 替换完就可以搜索下载了

2.5 mac\_lm-studio 换源

```
- 查找文件
grep -rl 'huggingface.co' /Applications/LM\ Studio.app/Contents/Resources/app/.webpack >/tmp/0224.f

igozhang@mac-m3 .webpack % grep -rl 'huggingface.co' /Applications/LM\ Studio.app/Contents/Resources/app/.webpack
/Applications/LM Studio.app/Contents/Resources/app/.webpack/renderer/main_window.js
/Applications/LM Studio.app/Contents/Resources/app/.webpack/renderer/main_window.js.map
/Applications/LM Studio.app/Contents/Resources/app/.webpack/lms
/Applications/LM Studio.app/Contents/Resources/app/.webpack/main/index.js

- 查找并输出到文件
while IFS= read -r file; do echo "$file"; sed -i '' 's/huggingface.co/hf-mirror.com/g' "$file" ;done < /tmp/0224.f
由于mac_SIP数据完整性保护，无法修改，加sudo也不行，开SIP步骤：
csrutil status 检查状态
要关闭这个才可以修改

由于编码，文件sip，文件路径空格等问题，快速手动替换完成，然后重启lm-studio就可以了
```
