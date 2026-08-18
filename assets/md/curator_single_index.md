# curator针对elasticsearch单条索引做生命周期策略

```
需求：
“igo_acc_log*”索引保留6个月，其他索引保留一个月；

ENV：
elasticsearch-7.9.3-1.x86_64
elasticsearch-curator-5.8.4-1.x86_64
```

```
配置（实测通过）
cat /opt/es_clean/delete_indices.yml
actions:
  1:
    action: delete_indices
    description: "Delete indices older than 6 months (for 'igo_acc_log*' only) or 1 month (for others)"
    options:
      ignore_empty_list: True
      timeout_override:
      continue_if_exception: False
    filters:
      - filtertype: pattern
        kind: prefix
        value: igo_acc_log*
        exclude: False
      - filtertype: age
        source: name
        direction: older
        timestring: '%Y.%m.%d'
        unit: days
        unit_count: 180
      - filtertype: pattern
        kind: prefix
        value: igo_acc_log*
        exclude: True
      - filtertype: age
        source: name
        direction: older
        timestring: '%Y.%m.%d'
        unit: days
        unit_count: 30

说明：（by阿里 通义千问https://qianwen.aliyun.com/ 赞）
在这个配置文件中，我们首先定义了一个名为`delete_indices`的Action，它会删除符合筛选条件的所有索引。然后我们定义了四个Filter：

- 第一个Filter是一个名为`pattern`的Filter，它匹配所有以`igo_acc_log*`开头的索引，并排除它们（因为后面我们会针对这些索引定义不同的保留策略）。这里我们将`exclude`设置为了`False`，表示要包含匹配到的索引。
- 第二个Filter是一个名为`age`的Filter，它会对索引名称进行过滤，并只保留那些创建时间不超过6个月（即满足`direction: older`和`unit_count: 180`条件）的索引。这里的`timestring`参数用于定义索引名中的时间格式，以便于计算索引的实际创建时间。
- 第三个Filter再次使用了`pattern` Filter，这次我们将其设置为排除所有以`igo_acc_log*`开头的索引（因为前面已经处理过了这些索引），并将`exclude`设置为了`True`。
- 最后一个Filter则定义了一种针对除`igo_acc_log*`以外所有其他索引的过期策略，即只保留最近1个月内（即满足`direction: older`和`unit_count: 30`条件）的索引。

这样，在运行这个配置文件时，Curator将会分别针对以`igo_acc_log*`开头的索引和其他索引执行不同的保留策略，即前者将保留最近6个月内的索引，而后者将保留最近1个月内的索引。
```
