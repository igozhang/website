# webhook

## 飞书

创建机器人  
复制webhook\_url  
测试

```
curl -X POST 'https://open.feishu.cn/open-apis/bot/v2/hook/e52319a0-09dc-40f1-8132-73d7df06aigo' \
  -H 'Content-Type: application/json' \
  -d '{"msg_type":"text","content":{"text":"hello"}}'

发个彩色表格
curl -X POST 'https://open.feishu.cn/open-apis/bot/v2/hook/e52319a0-09dc-40f1-8132-73d7df06aigo' \
  -H 'Content-Type: application/json' \
  -d '{
  "msg_type": "interactive",
  "card": {
    "config": { "wide_screen_mode": true },
    "header": {
      "title": { "tag": "plain_text", "content": "表格示例" },
      "template": "blue"
    },
    "elements": [
      {
        "tag": "table",
        "page_size": 5,
        "row_height": "low",
        "header_style": {
          "text_align": "left",
          "text_size": "normal",
          "background_style": "wathet",
          "text_color": "grey",
          "bold": true,
          "lines": 1
        },
        "columns": [
          { "name": "name", "display_name": "名称", "data_type": "text" },
          { "name": "level", "display_name": "级别", "data_type": "options" }
        ],
        "rows": [
          { "name": "飞书科技", "level": [{ "text": "S2", "color": "blue" }] },
          { "name": "飞书科技_01", "level": [{ "text": "S1", "color": "red" }] }
        ]
      }
    ]
  }
}'

OUTPUT
{"StatusCode":0,"StatusMessage":"success","code":0,"data":{},"msg":"success"}
```
