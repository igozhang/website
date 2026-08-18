# dify_flow

一个基于dify对话流的资源使用可视化  
https://cloud.dify.ai/app/97adc65f-effc-418a-9d5d-371dd5c1bab4/workflow

## 需求：

基于资源使用率数据，生成可视化图表，帮助快速发现问题（饼图，住图，折线图）

## 数据：

资源使用数据  
服务器IP cpu使用率P95  
192.168.1.1 13  
192.168.1.2 22  
192.168.1.3 10  
192.168.1.4 11  
192.168.1.5 8  
192.168.1.6 6  
192.168.1.7 31

## 可导入DSL\_YAML

```
app:
  description: 基于资源使用率数据，生成可视化图表，帮助快速发现问题（饼图，住图，折线图）
  icon: 🤖
  icon_background: '#FFEAD5'
  mode: advanced-chat
  name: AI画饼
  use_icon_as_answer_icon: false
dependencies:
- current_identifier: null
  type: marketplace
  value:
    marketplace_plugin_unique_identifier: langgenius/echarts:0.0.1@fe95412f101f2735530ea05f479ab5e8fbe1d4f4b2847f09e490e7d1d0125de5
- current_identifier: null
  type: marketplace
  value:
    marketplace_plugin_unique_identifier: langgenius/deepseek:0.0.5@fd6efd37c2a931911de8ab9ca3ba2da303bef146d45ee87ad896b04b36d09403
kind: app
version: 0.2.0
workflow:
  conversation_variables: []
  environment_variables: []
  features:
    file_upload:
      allowed_file_extensions:
      - .JPG
      - .JPEG
      - .PNG
      - .GIF
      - .WEBP
      - .SVG
      allowed_file_types:
      - image
      allowed_file_upload_methods:
      - local_file
      - remote_url
      enabled: false
      fileUploadConfig:
        audio_file_size_limit: 50
        batch_count_limit: 5
        file_size_limit: 15
        image_file_size_limit: 10
        video_file_size_limit: 100
        workflow_file_upload_limit: 10
      image:
        enabled: false
        number_limits: 3
        transfer_methods:
        - local_file
        - remote_url
      number_limits: 3
    opening_statement: ''
    retriever_resource:
      enabled: true
    sensitive_word_avoidance:
      enabled: false
    speech_to_text:
      enabled: false
    suggested_questions: []
    suggested_questions_after_answer:
      enabled: false
    text_to_speech:
      enabled: false
      language: ''
      voice: ''
  graph:
    edges:
    - data:
        isInLoop: false
        sourceType: tool
        targetType: answer
      id: 1743818987278-source-answer-target
      selected: false
      source: '1743818987278'
      sourceHandle: source
      target: answer
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: start
        targetType: llm
      id: 1743817081121-source-1743821229753-target
      selected: false
      source: '1743817081121'
      sourceHandle: source
      target: '1743821229753'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: llm
        targetType: if-else
      id: 1743821229753-source-1743821246429-target
      selected: false
      source: '1743821229753'
      sourceHandle: source
      target: '1743821246429'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: answer
      id: 1743821246429-false-1743821361833-target
      selected: false
      source: '1743821246429'
      sourceHandle: 'false'
      target: '1743821361833'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: document-extractor
      id: 1743821246429-true-1743821144786-target
      selected: false
      source: '1743821246429'
      sourceHandle: 'true'
      target: '1743821144786'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: document-extractor
        targetType: parameter-extractor
      id: 1743821144786-source-1743823668762-target
      selected: false
      source: '1743821144786'
      sourceHandle: source
      target: '1743823668762'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: parameter-extractor
        targetType: llm
      id: 1743823668762-source-1743824718722-target
      selected: false
      source: '1743823668762'
      sourceHandle: source
      target: '1743824718722'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: llm
        targetType: code
      id: 1743824718722-source-1743826125518-target
      source: '1743824718722'
      sourceHandle: source
      target: '1743826125518'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: code
        targetType: code
      id: 1743826125518-source-1743825188681-target
      source: '1743826125518'
      sourceHandle: source
      target: '1743825188681'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: code
        targetType: code
      id: 1743826125518-source-17438252572710-target
      source: '1743826125518'
      sourceHandle: source
      target: '17438252572710'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: code
        targetType: if-else
      id: 1743825188681-source-1743838183048-target
      source: '1743825188681'
      sourceHandle: source
      target: '1743838183048'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: code
        targetType: if-else
      id: 17438252572710-source-1743838183048-target
      source: '17438252572710'
      sourceHandle: source
      target: '1743838183048'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: tool
      id: 1743838183048-true-1743818987278-target
      source: '1743838183048'
      sourceHandle: 'true'
      target: '1743818987278'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: tool
      id: 1743838183048-8e8b889e-917b-448a-8bcf-804bc315ca4f-1743838165430-target
      source: '1743838183048'
      sourceHandle: 8e8b889e-917b-448a-8bcf-804bc315ca4f
      target: '1743838165430'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: tool
      id: 1743838183048-25edd05e-278b-47c8-86c6-6c6fe72ee69d-1743838365704-target
      source: '1743838183048'
      sourceHandle: 25edd05e-278b-47c8-86c6-6c6fe72ee69d
      target: '1743838365704'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: if-else
        targetType: answer
      id: 1743838183048-false-1743838388659-target
      source: '1743838183048'
      sourceHandle: 'false'
      target: '1743838388659'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: tool
        targetType: answer
      id: 1743838165430-source-1743838380498-target
      source: '1743838165430'
      sourceHandle: source
      target: '1743838380498'
      targetHandle: target
      type: custom
      zIndex: 0
    - data:
        isInLoop: false
        sourceType: tool
        targetType: answer
      id: 1743838365704-source-1743838384720-target
      source: '1743838365704'
      sourceHandle: source
      target: '1743838384720'
      targetHandle: target
      type: custom
      zIndex: 0
    nodes:
    - data:
        desc: ''
        selected: false
        title: 开始
        type: start
        variables:
        - allowed_file_extensions: []
          allowed_file_types:
          - document
          allowed_file_upload_methods:
          - local_file
          - remote_url
          label: 图表数据Excel
          max_length: 48
          options: []
          required: true
          type: file
          variable: excel
        - allowed_file_extensions: []
          allowed_file_types:
          - image
          allowed_file_upload_methods:
          - local_file
          - remote_url
          label: 图表类型
          max_length: 5
          options:
          - 柱状图
          - 饼状图
          - 折线图
          required: true
          type: select
          variable: chat_type
      height: 116
      id: '1743817081121'
      position:
        x: -56.474542032630495
        y: 304.89250382482834
      positionAbsolute:
        x: -56.474542032630495
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        answer: '{{#1743818987278.text#}}'
        desc: ''
        selected: false
        title: 直接回复 饼状图
        type: answer
        variables: []
      height: 105
      id: answer
      position:
        x: 1734.4142173884682
        y: 304.89250382482834
      positionAbsolute:
        x: 1734.4142173884682
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        desc: ''
        is_team_authorization: true
        output_schema: null
        paramSchemas:
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: chart title
            ja_JP: chart title
            pt_BR: chart title
            zh_Hans: 图表的标题
          label:
            en_US: title
            ja_JP: title
            pt_BR: title
            zh_Hans: 标题
          llm_description: chart title
          max: null
          min: null
          name: title
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: data for generating chart, each number should be separated by ";"
            ja_JP: data for generating chart, each number should be separated by ";"
            pt_BR: data for generating chart, each number should be separated by ";"
            zh_Hans: 用于生成饼图的数据，每个数字之间用 ";" 分隔
          label:
            en_US: data
            ja_JP: data
            pt_BR: data
            zh_Hans: 数据
          llm_description: data for generating pie chart, data should be a string
            contains a list of numbers like "1;2;3;4;5"
          max: null
          min: null
          name: data
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: Categories for chart, each category should be separated by ";"
            ja_JP: Categories for chart, each category should be separated by ";"
            pt_BR: Categories for chart, each category should be separated by ";"
            zh_Hans: 饼图的分类，每个分类之间用 ";" 分隔
          label:
            en_US: Categories
            ja_JP: Categories
            pt_BR: Categories
            zh_Hans: 分类
          llm_description: categories for pie chart, categories should be a string
            contains a list of texts like "a;b;c;1;2" in order to match the data,
            each category should be split by ";"
          max: null
          min: null
          name: categories
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        params:
          categories: ''
          data: ''
          title: ''
        provider_id: langgenius/echarts/echarts
        provider_name: langgenius/echarts/echarts
        provider_type: builtin
        selected: false
        title: 饼图
        tool_configurations: {}
        tool_label: 饼图
        tool_name: pie_chart
        tool_parameters:
          categories:
            type: mixed
            value: '{{#17438252572710.ip#}}'
          data:
            type: mixed
            value: '{{#1743825188681.data#}}'
          title:
            type: mixed
            value: '{{#1743823668762.title#}}'
        type: tool
      height: 54
      id: '1743818987278'
      position:
        x: 1408.4471731827994
        y: 304.89250382482834
      positionAbsolute:
        x: 1408.4471731827994
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        desc: ''
        is_array_file: false
        selected: false
        title: 文档提取器 提取Excel内容
        type: document-extractor
        variable_selector:
        - '1743817081121'
        - excel
      height: 92
      id: '1743821144786'
      position:
        x: 518.7107069021347
        y: 304.89250382482834
      positionAbsolute:
        x: 518.7107069021347
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        context:
          enabled: true
          variable_selector:
          - '1743817081121'
          - excel
          - extension
        desc: ''
        model:
          completion_params: {}
          mode: chat
          name: deepseek-chat
          provider: langgenius/deepseek/deepseek
        prompt_template:
        - id: b9b0b63c-09d9-4f99-aa15-27b16217791f
          role: system
          text: '【任务】

            请判断上下文这个文件的后缀名是否是Excel支持的文档格式，

            如果是，请回复数字 1

            如果不是，请回复数字 0

            【上下文】

            {{#context#}}'
        selected: false
        title: LLM
        type: llm
        variables: []
        vision:
          enabled: false
      height: 90
      id: '1743821229753'
      position:
        x: 220.96294397805582
        y: 304.89250382482834
      positionAbsolute:
        x: 220.96294397805582
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        cases:
        - case_id: 'true'
          conditions:
          - comparison_operator: is
            id: 36070cb9-2313-45bd-8151-aad9d5acba57
            value: '1'
            varType: string
            variable_selector:
            - '1743821229753'
            - text
          id: 'true'
          logical_operator: and
        desc: ''
        selected: false
        title: 条件分支 判断文档是否Excel
        type: if-else
      height: 126
      id: '1743821246429'
      position:
        x: 220.96294397805582
        y: 427.46429272661265
      positionAbsolute:
        x: 220.96294397805582
        y: 427.46429272661265
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        answer: <font color="red">请上传Excel格式的文档！</font>
        desc: ''
        selected: false
        title: 直接回复 不是Excel
        type: answer
        variables: []
      height: 118
      id: '1743821361833'
      position:
        x: 220.96294397805582
        y: 576.5736512188123
      positionAbsolute:
        x: 220.96294397805582
        y: 576.5736512188123
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        desc: ''
        instruction: '【任务】

          请根据输入的变量，提取数据

          【输入变量】

          {{#1743821144786.text#}}'
        model:
          completion_params:
            temperature: 0.7
          mode: chat
          name: deepseek-chat
          provider: langgenius/deepseek/deepseek
        parameters:
        - description: 销售数量
          name: counts
          required: true
          type: array[number]
        - description: IP
          name: ip
          required: true
          type: array[string]
        - description: 标题
          name: title
          required: false
          type: string
        query:
        - '1743821144786'
        - text
        reasoning_mode: prompt
        selected: false
        title: 参数提取器
        type: parameter-extractor
        variables: []
        vision:
          enabled: false
      height: 90
      id: '1743823668762'
      position:
        x: 518.7107069021347
        y: 427.46429272661265
      positionAbsolute:
        x: 518.7107069021347
        y: 427.46429272661265
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        context:
          enabled: false
          variable_selector: []
        desc: ''
        model:
          completion_params: {}
          mode: chat
          name: deepseek-chat
          provider: langgenius/deepseek/deepseek
        prompt_template:
        - id: 9e5152b1-7633-465c-9b95-ba0532a5b863
          role: system
          text: '【任务】

            我会提供参数供你格式化。每个参数都是一个Array的对象，把Array组装成用分号";"连接的字符串。最终把这些数据组装成示例中的json格式的字符串。请务必严格按照示例输出。

            【参数】

            {{#1743823668762.ip#}}

            {{#1743823668762.counts#}}

            【示例】

            {

            "ip": "ip1;ip2;ip3;ip4;ip5",

            "data": "10;20;30;40;50"

            }'
        selected: false
        title: LLM 数据JSON
        type: llm
        variables: []
        vision:
          enabled: false
      height: 90
      id: '1743824718722'
      position:
        x: 518.7107069021347
        y: 533.7738993792943
      positionAbsolute:
        x: 518.7107069021347
        y: 533.7738993792943
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        code: "\nfunction main({json}) {\n    return {\n        data: json.data\n\
          \    }\n}\n"
        code_language: javascript
        desc: ''
        outputs:
          data:
            children: null
            type: string
        selected: true
        title: 代码执行 获得使用率
        type: code
        variables:
        - value_selector:
          - '1743826125518'
          - result
          variable: json
      height: 54
      id: '1743825188681'
      position:
        x: 808.8977156541523
        y: 510.7765430309694
      positionAbsolute:
        x: 808.8977156541523
        y: 510.7765430309694
      selected: true
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        code: "\nfunction main({json}) {\n    return {\n        ip: json.ip\n    }\n\
          }\n"
        code_language: javascript
        desc: ''
        outputs:
          ip:
            children: null
            type: string
        selected: false
        title: 代码执行 获得IP
        type: code
        variables:
        - value_selector:
          - '1743826125518'
          - result
          variable: json
      height: 54
      id: '17438252572710'
      position:
        x: 808.8977156541523
        y: 414.6103360011026
      positionAbsolute:
        x: 808.8977156541523
        y: 414.6103360011026
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        code: "\nfunction main({str}) {\n    return {\n        result: JSON.parse(str)\n\
          \    }\n}\n"
        code_language: javascript
        desc: ''
        outputs:
          result:
            children: null
            type: object
        selected: false
        title: 代码执行 字符串转对象
        type: code
        variables:
        - value_selector:
          - '1743824718722'
          - text
          variable: str
      height: 54
      id: '1743826125518'
      position:
        x: 808.8977156541523
        y: 304.89250382482834
      positionAbsolute:
        x: 808.8977156541523
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        desc: ''
        is_team_authorization: true
        output_schema: null
        paramSchemas:
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: chart title
            ja_JP: chart title
            pt_BR: chart title
            zh_Hans: 图表的标题
          label:
            en_US: title
            ja_JP: title
            pt_BR: title
            zh_Hans: 标题
          llm_description: chart title
          max: null
          min: null
          name: title
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: data for generating chart, each number should be separated by ";"
            ja_JP: data for generating chart, each number should be separated by ";"
            pt_BR: data for generating chart, each number should be separated by ";"
            zh_Hans: 用于生成柱状图的数据，每个数字之间用 ";" 分隔
          label:
            en_US: data
            ja_JP: data
            pt_BR: data
            zh_Hans: 数据
          llm_description: data for generating bar chart, data should be a string
            contains a list of numbers like "1;2;3;4;5"
          max: null
          min: null
          name: data
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: X axis for chart, each text should be separated by ";"
            ja_JP: X axis for chart, each text should be separated by ";"
            pt_BR: X axis for chart, each text should be separated by ";"
            zh_Hans: 柱状图的 x 轴，每个文本之间用 ";" 分隔
          label:
            en_US: X Axis
            ja_JP: X Axis
            pt_BR: X Axis
            zh_Hans: x 轴
          llm_description: x axis for bar chart, x axis should be a string contains
            a list of texts like "a;b;c;1;2" in order to match the data
          max: null
          min: null
          name: x_axis
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        params:
          data: ''
          title: ''
          x_axis: ''
        provider_id: langgenius/echarts/echarts
        provider_name: langgenius/echarts/echarts
        provider_type: builtin
        selected: false
        title: 柱状图
        tool_configurations: {}
        tool_label: 柱状图
        tool_name: bar_chart
        tool_parameters:
          data:
            type: mixed
            value: '{{#1743825188681.data#}}'
          title:
            type: mixed
            value: '{{#1743823668762.title#}}'
          x_axis:
            type: mixed
            value: '{{#17438252572710.ip#}}'
        type: tool
      height: 54
      id: '1743838165430'
      position:
        x: 1408.4471731827994
        y: 392.79761321086556
      positionAbsolute:
        x: 1408.4471731827994
        y: 392.79761321086556
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        cases:
        - case_id: 'true'
          conditions:
          - comparison_operator: is
            id: 46ad82e9-744c-4e4e-91e9-2784020e03c6
            value: 饼状图
            varType: string
            variable_selector:
            - '1743817081121'
            - chat_type
          id: 'true'
          logical_operator: and
        - case_id: 8e8b889e-917b-448a-8bcf-804bc315ca4f
          conditions:
          - comparison_operator: is
            id: e595c9ff-e60c-4f74-924e-68f51ca22ff8
            value: 柱状图
            varType: string
            variable_selector:
            - '1743817081121'
            - chat_type
          id: 8e8b889e-917b-448a-8bcf-804bc315ca4f
          logical_operator: and
        - case_id: 25edd05e-278b-47c8-86c6-6c6fe72ee69d
          conditions:
          - comparison_operator: is
            id: bdf6d315-509d-4b8a-ab41-84f9f307a4f9
            value: 折线图
            varType: string
            variable_selector:
            - '1743817081121'
            - chat_type
          id: 25edd05e-278b-47c8-86c6-6c6fe72ee69d
          logical_operator: and
        desc: ''
        selected: false
        title: 条件分支 判断图表类型
        type: if-else
      height: 222
      id: '1743838183048'
      position:
        x: 1103.7884425496106
        y: 304.89250382482834
      positionAbsolute:
        x: 1103.7884425496106
        y: 304.89250382482834
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        desc: ''
        is_team_authorization: true
        output_schema: null
        paramSchemas:
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: chart title
            ja_JP: chart title
            pt_BR: chart title
            zh_Hans: 图表的标题
          label:
            en_US: title
            ja_JP: title
            pt_BR: title
            zh_Hans: 标题
          llm_description: chart title
          max: null
          min: null
          name: title
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: data for generating chart, each number should be separated by ";"
            ja_JP: data for generating chart, each number should be separated by ";"
            pt_BR: data for generating chart, each number should be separated by ";"
            zh_Hans: 用于生成线性图表的数据，每个数字之间用 ";" 分隔
          label:
            en_US: data
            ja_JP: data
            pt_BR: data
            zh_Hans: 数据
          llm_description: data for generating linear chart, data should be a string
            contains a list of numbers like "1;2;3;4;5"
          max: null
          min: null
          name: data
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        - auto_generate: null
          default: null
          form: llm
          human_description:
            en_US: X axis for chart, each text should be separated by ";"
            ja_JP: X axis for chart, each text should be separated by ";"
            pt_BR: X axis for chart, each text should be separated by ";"
            zh_Hans: 线性图表的 x 轴，每个文本之间用 ";" 分隔
          label:
            en_US: X Axis
            ja_JP: X Axis
            pt_BR: X Axis
            zh_Hans: x 轴
          llm_description: x axis for linear chart, x axis should be a string contains
            a list of texts like "a;b;c;1;2" in order to match the data
          max: null
          min: null
          name: x_axis
          options: []
          placeholder: null
          precision: null
          required: true
          scope: null
          template: null
          type: string
        params:
          data: ''
          title: ''
          x_axis: ''
        provider_id: langgenius/echarts/echarts
        provider_name: langgenius/echarts/echarts
        provider_type: builtin
        selected: false
        title: 线性图表
        tool_configurations: {}
        tool_label: 线性图表
        tool_name: line_chart
        tool_parameters:
          data:
            type: mixed
            value: '{{#1743825188681.data#}}'
          title:
            type: mixed
            value: '{{#1743823668762.title#}}'
          x_axis:
            type: mixed
            value: '{{#17438252572710.ip#}}'
        type: tool
      height: 54
      id: '1743838365704'
      position:
        x: 1408.4471731827994
        y: 470.1435319804011
      positionAbsolute:
        x: 1408.4471731827994
        y: 470.1435319804011
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        answer: '{{#1743838165430.text#}}'
        desc: ''
        selected: false
        title: 直接回复 柱状图
        type: answer
        variables: []
      height: 105
      id: '1743838380498'
      position:
        x: 1734.4142173884682
        y: 427.46429272661265
      positionAbsolute:
        x: 1734.4142173884682
        y: 427.46429272661265
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        answer: '{{#1743838365704.text#}}'
        desc: ''
        selected: false
        title: 直接回复 折线图
        type: answer
        variables: []
      height: 105
      id: '1743838384720'
      position:
        x: 1734.4142173884682
        y: 538.5445508034148
      positionAbsolute:
        x: 1734.4142173884682
        y: 538.5445508034148
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    - data:
        answer: <font color="orange">没有符合类型的图标</font>
        desc: ''
        selected: false
        title: 直接回复 没有图表类型
        type: answer
        variables: []
      height: 118
      id: '1743838388659'
      position:
        x: 1103.7884425496106
        y: 562.9243574337557
      positionAbsolute:
        x: 1103.7884425496106
        y: 562.9243574337557
      selected: false
      sourcePosition: right
      targetPosition: left
      type: custom
      width: 244
    viewport:
      x: 83.96320577722508
      y: -54.3340624762277
      zoom: 0.7455430912233502
```
