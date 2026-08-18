# shell_sort

```
用shell语言实现插入排序，希尔排序，冒泡排序，堆排序，选择排序，快速排序，归并排序算法并注释说明时间复杂度及原理
```

```
插入排序：
# 时间复杂度：O(n^2)
# 原理：插入排序是一种简单的排序算法，它的工作原理是通过构建有序序列，对于未排序数据，在已排序序列中从后向前扫描，找到相应位置并插入。
#!/bin/bash
# 插入排序
insertion_sort() {
    arr=($@)
    for ((i=1; i<${#arr[@]}; i++))
    do
        tmp=${arr[$i]}
        j=$(($i-1))
        while [ $j -ge 0 -a ${arr[$j]} -gt $tmp ]
        do
            arr[$(($j+1))]=${arr[$j]}
            j=$(($j-1))
        done
        arr[$(($j+1))]=$tmp
    done
    echo ${arr[@]}
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(insertion_sort ${arr[@]})"
希尔排序：
# 时间复杂度：O(n^1.3)
# 原理：希尔排序是把记录按下标的一定增量分组，对每组使用直接插入排序算法排序；随着增量逐渐减少，每组包含的关键词越来越多，当增量减至1时，整个文件恰被分成一组，算法便终止。
#!/bin/bash
# 希尔排序
shell_sort() {
    arr=($@)
    gap=${#arr[@]}
    while [ $((gap/=2)) -gt 0 ]
    do
        for ((i=$gap; i<${#arr[@]}; i++))
        do
            tmp=${arr[$i]}
            j=$(($i-$gap))
            while [ $j -ge 0 -a ${arr[$j]} -gt $tmp ]
            do
                arr[$(($j+$gap))]=${arr[$j]}
                j=$(($j-$gap))
            done
            arr[$(($j+$gap))]=$tmp
        done
    done
    echo ${arr[@]}
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(shell_sort ${arr[@]})"
冒泡排序：
# 时间复杂度：O(n^2)
# 原理：冒泡排序是一种简单的排序算法，它重复地走访过要排序的数列，一次比较两个元素，如果他们的顺序错误就把他们交换过来。
#!/bin/bash
# 冒泡排序
bubble_sort() {
    arr=($@)
    for ((i=0; i<${#arr[@]}; i++))
    do
        for ((j=0; j<$((${#arr[@]}-i-1)); j++))
        do
            if [ ${arr[$j]} -gt ${arr[$(($j+1))]} ]
            then
                tmp=${arr[$j]}
                arr[$j]=${arr[$(($j+1))]}
                arr[$(($j+1))]=$tmp
            fi
        done
    done
    echo ${arr[@]}
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(bubble_sort ${arr[@]})"
堆排序：
# 时间复杂度：O(nlogn)
# 原理：堆排序是一种树形选择排序，是对直接选择排序的有效改进。堆的定义如下：具有n个元素的序列（h1,h2,…,hn),当且仅当满足（hi>=h2i,hi>=2i+1）或（hi<=h2i,hi<=2i+1）(i=1,2,…,n/2)时称之为堆。
#!/bin/bash
# 堆排序
heap_sort() {
    arr=($@)
    # 建堆
    for ((i=${#arr[@]}/2-1; i>=0; i--))
    do
        heapify $i ${#arr[@]}
    done
    # 堆排序
    for ((i=${#arr[@]}-1; i>=0; i--))
    do
        tmp=${arr[0]}
        arr[0]=${arr[$i]}
        arr[$i]=$tmp
        heapify 0 $i
    done
    echo ${arr[@]}
}
# 堆调整
heapify() {
    start=$1
    end=$2
    root=${arr[$start]}
    while [ $(($start*2+1)) -lt $end ]
    do
        child=$(($start*2+1))
        if [ $child -lt $(($end-1)) -a ${arr[$child]} -lt ${arr[$(($child+1))]} ]
        then
            child=$(($child+1))
        fi
        if [ $root -lt ${arr[$child]} ]
        then
            arr[$start]=${arr[$child]}
            start=$child
        else
            break
        fi
    done
    arr[$start]=$root
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(heap_sort ${arr[@]})"
选择排序：
# 时间复杂度：O(n^2)
# 原理：选择排序是一种简单直观的排序算法，它的工作原理是每一次从待排序的数据元素中选出最小（或最大）的一个元素，存放在序列的起始位置，直到全部待排序的数据元素排完。
#!/bin/bash
# 选择排序
selection_sort() {
    arr=($@)
    for ((i=0; i<${#arr[@]}; i++))
    do
        min=$i
        for ((j=$(($i+1)); j<${#arr[@]}; j++))
        do
            if [ ${arr[$j]} -lt ${arr[$min]} ]
            then
                min=$j
            fi
        done
        tmp=${arr[$i]}
        arr[$i]=${arr[$min]}
        arr[$min]=$tmp
    done
    echo ${arr[@]}
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(selection_sort ${arr[@]})"
快速排序：
# 时间复杂度：O(nlogn)
# 原理：快速排序是对冒泡排序的一种改进，它的基本思想是：通过一趟排序将要排序的数据分割成独立的两部分，其中一部分的所有数据都比另外一部分的所有数据都要小，然后再按此方法对这两部分数据分别进行快速排序，整个排序过程可以递归进行，以此达到整个数据变成有序序列。
#!/bin/bash
# 快速排序
quick_sort() {
    arr=($@)
    __quick_sort 0 $((${#arr[@]}-1))
    echo ${arr[@]}
}
# 快速排序递归函数
__quick_sort() {
    start=$1
    end=$2
    if [ $start -lt $end ]
    then
        i=$start
        j=$end
        base=${arr[$i]}
        while [ $i -lt $j ]
        do
            while [ $i -lt $j -a ${arr[$j]} -ge $base ]
            do
                j=$(($j-1))
            done
            arr[$i]=${arr[$j]}
            while [ $i -lt $j -a ${arr[$i]} -le $base ]
            do
                i=$(($i+1))
            done
            arr[$j]=${arr[$i]}
        done
        arr[$i]=$base
        __quick_sort $start $(($i-1))
        __quick_sort $(($i+1)) $end
    fi
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(quick_sort ${arr[@]})"
归并排序：
# 时间复杂度：O(nlogn)
# 原理：归并排序是建立在归并操作上的一种有效的排序算法，该算法是采用分治法（Divide and Conquer）的一个非常典型的应用。将已有序的子序列合并，得到完全有序的序列；即先使每个子序列有序，再使子序列段间有序。若将两个有序表合并成一个有序表，称为2-路归并。
#!/bin/bash
# 归并排序
merge_sort() {
    arr=($@)
    __merge_sort 0 $((${#arr[@]}-1))
    echo ${arr[@]}
}
# 归并排序递归函数
__merge_sort() {
    start=$1
    end=$2
    if [ $start -lt $end ]
    then
        mid=$((($start+$end)/2))
        __merge_sort $start $mid
        __merge_sort $(($mid+1)) $end
        __merge $start $mid $end
    fi
}
# 合并函数
__merge() {
    start=$1
    mid=$2
    end=$3
    i=$start
    j=$(($mid+1))
    k=0
    tmp=()
    while [ $i -le $mid -a $j -le $end ]
    do
        if [ ${arr[$i]} -lt ${arr[$j]} ]
        then
            tmp[$k]=${arr[$i]}
            i=$(($i+1))
        else
            tmp[$k]=${arr[$j]}
            j=$(($j+1))
        fi
        k=$(($k+1))
    done
    while [ $i -le $mid ]
    do
        tmp[$k]=${arr[$i]}
        i=$(($i+1))
        k=$(($k+1))
    done
    while [ $j -le $end ]
    do
        tmp[$k]=${arr[$j]}
        j=$(($j+1))
        k=$(($k+1))
    done
    for ((i=$start; i<=$end; i++))
    do
        arr[$i]=${tmp[$(($i-$start))]}
    done
}
arr=(5 4 3 2 1)
echo "排序前：${arr[@]}"
echo "排序后：$(merge_sort ${arr[@]})"
```
