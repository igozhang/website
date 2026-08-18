# csv2json

```
通过python将指定字符串csv数据转换成json数据
```

```
# -*- coding: utf-8 -*-

import json

def main():
    # 原始文本数据
    data = """
零售,零售业务,RMS,零售管理系统,门店运营平台,RMS_StoreOps,store-core,retail,sales-service,store-core.retail.sales-service,张三
金融,金融服务,FMS,金融管理系统,信贷审批平台,FMS_CreditApprove,finance-core,lending,loan-engine,finance-core.lending.loan-engine,李四
物流,供应链管理,LMS,物流管理系统,智能调度系统,LMS_SmartDispatch,logistics-core,fleet,routing-engine,logistics-core.fleet.routing-engine,王五
制造,智能制造,MES,生产执行系统,工厂监控平台,MES_PlantMonitor,manufacture-core,plant,monitor-service,manufacture-core.plant.monitor-service,赵六
电商,电子商务,ECS,电商平台系统,订单处理中心,ECS_OrderCenter,ecommerce-core,order,order-processing,ecommerce-core.order.order-processing,陈七
""".strip()

    # 表头字段名
    headers = [
        "领域", "领域中文名称", "产品", "产品中文名称", "系统中文名称", "系统编码",
        "P", "S", "A", "PSA", "系统运维"
    ]

    # 解析数据行
    rows = data.split('\n')
    result = []
    for row in rows:
        fields = row.split(',')
        item = dict(zip(headers, fields))
        result.append(item)

    # 返回标准 JSON 字典格式
    return {
        "data": result
    }

if __name__ == "__main__":
    output = main()
    print(json.dumps(output, ensure_ascii=False, indent=2))
```

```
##将csv文件格式化并写入json文件python3代码

# -*- coding: utf-8 -*-

import csv
import json

def main():
    input_file = 'data.csv'
    output_file = 'cmdb01.json'

    result = []

    # 使用 utf-8-sig 自动忽略 BOM
    with open(input_file, mode='r', encoding='utf-8-sig') as csvfile:
        reader = csv.DictReader(csvfile)

        for row in reader:
            result.append(row)

    # 写入 JSON 文件
    with open(output_file, mode='w', encoding='utf-8') as jsonfile:
        json.dump(result, jsonfile, ensure_ascii=False, indent=2)

    print("✅ 已成功将 {} 条记录写入 {}".format(len(result), output_file))

if __name__ == "__main__":
    main()
```
