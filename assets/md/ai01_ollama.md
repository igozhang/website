# ai01_ollama

## 使用ollama再mac\_M3 上部署deepseek1.5b

ollama version is 0.5.11

tips: 除了ollma,LM Studio,vLLM,Chatbox等

对比大致如下(不完整)：  
Ollama : 专注于本地部署,term交互  
LM Studio:完整本地模型交互体验，支持联网，支持自检测硬件适配,有图形化界面  
Chatbox: 本地客户段，可以通过插件支持一些扩展调整，支持联网，但联网不如Chrome插件page\_assist好用  
Docker: docker部署open webui

另外page\_assist还可以帮助本地ollma部署的deepseek实现轻松联网

机器: M3 air 16G512G  
网址: ollama.com  
下载安装

1. deepseek 1.5b

```
sudo ollama run deepseek-r1:1.5b

>>> 你今天吃猪脚饭了吗，用英语怎么说
<think>

</think>

Hi! Yes, I have the opportunity to eat pig foot rice with you in an English-style 
bowl. Here it is: 

Pig foot rice is a traditional dish from China, and I am happy to share it with 
you. How much would you like? 😊

>>> /bye
igozhang@mac-m3 ~ %
```

2. deepseek 14b

```
igozhang@mac-m3 ~ % sudo ollama run deepseek-r1:14b
Password:
pulling manifest 
pulling 6e9f90f02bb3... 100% ▕███████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████▏ 9.0 GB                         
pulling 369ca498f347... 100% ▕███████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████▏  387 B                         
pulling 6e4c38e1172f... 100% ▕███████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████▏ 1.1 KB                         
pulling f4d24e9138dd... 100% ▕███████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████▏  148 B                         
pulling 3c24b0c80794... 100% ▕███████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████▏  488 B                         
verifying sha256 digest 
writing manifest 
success 
>>> 你会说英文吗？
<think>

</think>

Yes, I can communicate in English! How can I assist you today?

>>> Reproducibility of the model output 翻译这句话
<think>
Alright, let's take a look at this user query. The user wrote "Reproducibility of the model output 翻译这句话". So they want me to translate that specific phrase into Chinese.

First off, I need to understand each part of the phrase. "Reproducibility" refers to how reproducible something is. In this context, it's about the model output being reproducible. So, translating "reproducibility" accurately is key 
here.

I think "可重复性" is the right term for "reproducibility" in Chinese. It captures the meaning well without sounding too literal or awkward.

Next, "model output". This refers to the results produced by a model. In Chinese, that's best translated as "模型输出". Using "输出" instead of something like "结果" makes it clear that we're talking about what the model produces 
specifically.

Putting it all together, "Reproducibility of the model output" translates to "模型输出的可重复性". This translation feels natural and precise. I don't see any room for improvement here; it's straightforward and maintains the original 
meaning accurately.
</think>

模型输出的可重复性

>>> /bye
igozhang@mac-m3 ~ %
```

## 其他版本

```
ollama run deepseek-r1:1.5b  1.1GB
ollama run deepseek-r1:7b    4.7GB
ollama run deepseek-r1:8b    4.9GB
ollama run deepseek-r1:14b   9.0GB
ollama run deepseek-r1:32b   20GB
ollama run deepseek-r1:70b   43GB
ollama run deepseek-r1:70b   404GB

ollama run llama2:7b
ollama pull llama3:8b      4.7GB
ollama run llama3:70b
ollama pull llama3:70b --quantization=4bit
##Ollama 支持量化技术（如 4-bit 量化），可显著降低显存需求。例如
ollama run falcon:7b

ollama run qwen2
ollama run qwen2.5         4.7GB
2.5是语言模型，其他的还有数学模型Qwen2.5-Math，编程模型Qwen2.5-Coder，视觉模型Qwen2.5-VL等

删除命令
sudo ollama rm deepseek-r1:32b

实测M3 air 16G512G 跑deepseek1.5b很顺畅，14b比较顺畅，最多跑到32b 版本，但是很卡，蹦一个字可能都两三分钟
```

资源是动态加载的：在你同时安装了几个大模型的情况下  
动态加载：通常情况下，这些模型不会同时“时刻”运行。大多数平台或服务会采用按需加载的方式，即只有在收到请求时才会启动相应的模型进行处理。

## 3张4070如何同时跑deepseek

### 简单理解

PyTorch 提供将模型并行跑在多个显卡的方法  
torch.nn.DataParallel 是 PyTorch 提供的一个简单的多 GPU 并行计算工具，可以将模型分割复制到多个 GPU 上并行处理数据。  
还有其他方法可以指定调度显卡

torch.nn.DistributedDataParallel  
DistributedDataParallel 是更高效的多 GPU 并行方式，它通过分布式训练的机制，每个 GPU 独立计算梯度并进行同步，减少了主 GPU 的负担。

### 详细步骤

显卡驱动：安装最新的 NVIDIA 显卡驱动，以支持多 GPU 并行计算。你可以从 NVIDIA 官方网站 下载并安装适合你系统的驱动程序。  
CUDA 工具包：DeepSeek 模型通常基于 PyTorch 运行，而 PyTorch 需要 CUDA 支持。下载并安装与你的显卡驱动兼容的 CUDA 版本，可从 NVIDIA CUDA 官方下载页面 获取。  
cuDNN：cuDNN 是 NVIDIA 提供的深度神经网络库，可以加速深度学习模型的训练和推理。下载并安装与你 CUDA 版本匹配的 cuDNN，从 NVIDIA cuDNN 官方下载页面 获取。  
Python 和 PyTorch：安装 Python 环境（建议使用 Python 3.8 及以上版本），并安装支持 CUDA 的 PyTorch 版本。可以使用以下命令安装：  
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118

## 概念

1. 深度学习框架  
   PyTorch  
   基于 Python：PyTorch 构建在 Python 之上，充分利用了 Python 的简洁性和灵活性，使得开发者可以使用熟悉的 Python 语法来构建和训练深度学习模型，降低了学习成本，同时也方便与其他 Python 库集成。

TensorFlow  
动态计算图：与一些传统的深度学习框架（如 TensorFlow 1.x）使用静态计算图不同，PyTorch 采用动态计算图机制。这意味着在运行时可以动态地定义和修改计算图，使得代码的调试和开发更加直观和灵活，尤其适合研究人员进行快速迭代和实验。
