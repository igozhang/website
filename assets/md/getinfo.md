# getinfo

```
从一段文本中提取IPV4地址,或者域名等信息
```

```
import re

def extract_ip(text):
    """
    提取文本中的第一个 IPv4 地址
    """
    # 去掉 \b 边界，支持IP出现在中文等非单词字符之间
    ip_pattern = r'(?:[0-9]{1,3}\.){3}[0-9]{1,3}'
    match = re.search(ip_pattern, text)
    if match:
        return match.group(0)
    else:
        return "No IP address found"

def main(input_text):
    result = extract_ip(input_text)
    return {"sysip": result}
```

```
提取域名：
```

```
import re

def extract_domain(text):
    # 正则表达式匹配第一个出现的域名
    domain_pattern = r'(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}(?:\.[a-zA-Z]{2,})?'
    match = re.search(domain_pattern, text)
    return match.group(0) if match else None

def main(input_text):
    result = extract_domain(input_text)
    return {"sysdomain": result}
```
