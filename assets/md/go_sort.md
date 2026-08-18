# go_sort

```
用go语言实现插入排序，希尔排序，冒泡排序，堆排序，选择排序，快速排序，归并排序算法并注释说明时间复杂度及原理
```

```
插入排序：
// 时间复杂度：O(n^2)
// 原理：插入排序是一种简单的排序算法，它的工作原理是通过构建有序序列，对于未排序数据，在已排序序列中从后向前扫描，找到相应位置并插入。
func InsertionSort(arr []int) {
	for i := 1; i < len(arr); i++ {
		for j := i; j > 0; j-- {
			if arr[j] < arr[j-1] {
				arr[j], arr[j-1] = arr[j-1], arr[j]
			}
		}
	}
}
希尔排序：
// 时间复杂度：O(n^1.3)
// 原理：希尔排序是把记录按下标的一定增量分组，对每组使用直接插入排序算法排序；随着增量逐渐减少，每组包含的关键词越来越多，当增量减至1时，整个文件恰被分成一组，算法便终止。
func ShellSort(arr []int) {
	n := len(arr)
	h := 1
	for h < n/3 {
		h = 3*h + 1
	}
	for h >= 1 {
		for i := h; i < n; i++ {
			for j := i; j >= h && arr[j] < arr[j-h]; j -= h {
				arr[j], arr[j-h] = arr[j-h], arr[j]
			}
		}
		h /= 3
	}
}
冒泡排序：
// 时间复杂度：O(n^2)
// 原理：冒泡排序是一种简单的排序算法，它重复地走访过要排序的数列，一次比较两个元素，如果他们的顺序错误就把他们交换过来。
func BubbleSort(arr []int) {
	for i := 0; i < len(arr); i++ {
		for j := 0; j < len(arr)-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}
堆排序：
// 时间复杂度：O(nlogn)
// 原理：堆排序是一种树形选择排序，是对直接选择排序的有效改进。堆的定义如下：具有n个元素的序列（h1,h2,...,hn),当且仅当满足（hi>=h2i,hi>=2i+1）或（hi<=h2i,hi<=2i+1）(i=1,2,...,n/2)时称之为堆。
func HeapSort(arr []int) {
	n := len(arr)
	// 构建大顶堆
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, i, n)
	}
	// 堆排序
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heapify(arr, 0, i)
	}
}
func heapify(arr []int, i, n int) {
	for {
		maxPos := i
		if i*2+1 < n && arr[i] < arr[i*2+1] {
			maxPos = i*2 + 1
		}
		if i*2+2 < n && arr[maxPos] < arr[i*2+2] {
			maxPos = i*2 + 2
		}
		if maxPos == i {
			break
		}
		arr[i], arr[maxPos] = arr[maxPos], arr[i]
		i = maxPos
	}
}
选择排序：
// 时间复杂度：O(n^2)
// 原理：选择排序是一种简单直观的排序算法，它的工作原理是每一次从待排序的数据元素中选出最小（或最大）的一个元素，存放在序列的起始位置，直到全部待排序的数据元素排完。
func SelectionSort(arr []int) {
	for i := 0; i < len(arr); i++ {
		minIndex := i
		for j := i + 1; j < len(arr); j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}
		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}
}
快速排序：
// 时间复杂度：O(nlogn)
// 原理：快速排序是对冒泡排序的一种改进，它的基本思想是：通过一趟排序将要排序的数据分割成独立的两部分，其中一部分的所有数据都比另外一部分的所有数据都要小，然后再按此方法对这两部分数据分别进行快速排序，整个排序过程可以递归进行，以此达到整个数据变成有序序列。
func QuickSort(arr []int) {
	quickSort(arr, 0, len(arr)-1)
}
func quickSort(arr []int, left, right int) {
	if left < right {
		pivot := partition(arr, left, right)
		quickSort(arr, left, pivot-1)
		quickSort(arr, pivot+1, right)
	}
}
func partition(arr []int, left, right int) int {
	pivot := arr[right]
	i := left
	for j := left; j < right; j++ {
		if arr[j] < pivot {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[right] = arr[right], arr[i]
	return i
}
归并排序：
// 时间复杂度：O(nlogn)
// 原理：归并排序是建立在归并操作上的一种有效的排序算法，该算法是采用分治法（Divide and Conquer）的一个非常典型的应用。将已有序的子序列合并，得到完全有序的序列；即先使每个子序列有序，再使子序列段间有序。若将两个有序表合并成一个有序表，称为2-路归并。
func MergeSort(arr []int) {
	mergeSort(arr, 0, len(arr)-1)
}
func mergeSort(arr []int, left, right int) {
	if left < right {
		mid := (left + right) / 2
		mergeSort(arr, left, mid)
		mergeSort(arr, mid+1, right)
		merge(arr, left, mid, right)
	}
}
func merge(arr []int, left, mid, right int) {
	tmp := make([]int, right-left+1)
	i := left
	j := mid + 1
	k := 0
	for i <= mid && j <= right {
		if arr[i] <= arr[j] {
			tmp[k] = arr[i]
			i++
			k++
		} else {
			tmp[k] = arr[j]
			j++
			k++
		}
	}
	for i <= mid {
		tmp[k] = arr[i]
		i++
		k++
	}
	for j <= right {
		tmp[k] = arr[j]
		j++
		k++
	}
	copy(arr[left:right+1], tmp)
}
```
