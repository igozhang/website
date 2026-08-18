# java程序性能分析

java-analyse-go

# env

CentOS Linux release 7.9.2009 (Core)  
go version go1.20.12 linux/amd64  
openjdk version "1.8.0\_412"

输入： openjdk的java进程id  
输出： top3cpu线程，top3mem线程

历次优化说明：

1. 线程信息添加堆栈信息，便于分析
2. 内存单位展示优化，大于1G按GB显示，小于1G展示MB
3. 内存取数逻辑优化

# code

```
package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
)

// ThreadInfo 存储线程的 CPU、内存和堆栈信息
type ThreadInfo struct {
    TID    string
    CPU    float64
    Memory float64
    Stack  string
}

// GetProcessThreads 获取指定进程的所有线程 ID
func GetProcessThreads(pid string) ([]string, error) {
    threadsPath := filepath.Join("/proc", pid, "task")
    entries, err := os.ReadDir(threadsPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read threads directory: %v", err)
    }

    var threads []string
    for _, entry := range entries {
        if entry.IsDir() {
            threads = append(threads, entry.Name())
        }
    }
    return threads, nil
}

// ReadThreadStat 获取线程的 CPU 使用率
func ReadThreadStat(pid, tid string) (float64, error) {
    statPath := filepath.Join("/proc", pid, "task", tid, "stat")
    file, err := os.Open(statPath)
    if err != nil {
        return 0, fmt.Errorf("failed to open stat file: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    if scanner.Scan() {
        fields := strings.Fields(scanner.Text())
        if len(fields) < 14 {
            return 0, fmt.Errorf("stat file format invalid")
        }

        utime, err := strconv.ParseFloat(fields[13], 64) // 用户态时间
        if err != nil {
            return 0, fmt.Errorf("failed to parse utime: %v", err)
        }

        stime, err := strconv.ParseFloat(fields[14], 64) // 内核态时间
        if err != nil {
            return 0, fmt.Errorf("failed to parse stime: %v", err)
        }

        // 返回 CPU 使用时间
        return utime + stime, nil
    }
    return 0, fmt.Errorf("failed to read stat file")
}

// ReadThreadMemory 获取线程的实际驻留内存（RSS），以 GB 为单位
func ReadThreadMemory(pid, tid string) (float64, error) {
    statusPath := filepath.Join("/proc", pid, "task", tid, "status")
    file, err := os.Open(statusPath)
    if err != nil {
        return 0, fmt.Errorf("failed to open status file: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "VmRSS:") {
            fields := strings.Fields(line)
            if len(fields) < 2 {
                return 0, fmt.Errorf("invalid VmRSS line format")
            }

            // 将 VmRSS 值转换为浮点数（单位：kB）
            rssKb, err := strconv.ParseFloat(fields[1], 64)
            if err != nil {
                return 0, fmt.Errorf("failed to parse VmRSS value: %v", err)
            }

            // 转换为 GB
            return rssKb / (1024 * 1024), nil
        }
    }
    return 0, fmt.Errorf("VmRSS not found in status file")
}

// ReadThreadStack 获取线程的堆栈信息
func ReadThreadStack(pid, tid string) (string, error) {
    stackPath := filepath.Join("/proc", pid, "task", tid, "stack")
    file, err := os.Open(stackPath)
    if err != nil {
        return "", fmt.Errorf("failed to open stack file: %v", err)
    }
    defer file.Close()

    stackData, err := io.ReadAll(file)
    if err != nil {
        return "", fmt.Errorf("failed to read stack file: %v", err)
    }

    return string(stackData), nil
}

// FormatMemory 格式化内存大小，根据大小自动选择单位（GB 或 MB）
func FormatMemory(memory float64) string {
    if memory >= 1 {
        return fmt.Sprintf("%.2fGB", memory)
    }
    return fmt.Sprintf("%.2fMB", memory*1024)
}

// min 返回两个整数中的最小值
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: go run main.go <pid>")
        return
    }

    pid := os.Args[1]

    // 获取线程信息
    threads, err := GetProcessThreads(pid)
    if err != nil {
        fmt.Printf("Error getting threads: %v\n", err)
        return
    }

    var threadInfos []ThreadInfo
    for _, tid := range threads {
        cpu, err := ReadThreadStat(pid, tid)
        if err != nil {
            fmt.Printf("Error reading CPU info for thread %s: %v\n", tid, err)
            continue
        }

        mem, err := ReadThreadMemory(pid, tid)
        if err != nil {
            fmt.Printf("Error reading memory info for thread %s: %v\n", tid, err)
            continue
        }

        stack, err := ReadThreadStack(pid, tid)
        if err != nil {
            fmt.Printf("Error reading stack info for thread %s: %v\n", tid, err)
            continue
        }

        threadInfos = append(threadInfos, ThreadInfo{
            TID:    tid,
            CPU:    cpu,
            Memory: mem,
            Stack:  stack,
        })
    }

    // 按 CPU 排序
    sort.Slice(threadInfos, func(i, j int) bool {
        return threadInfos[i].CPU > threadInfos[j].CPU
    })
    fmt.Println("Top 3 threads by CPU usage:")
    for i, info := range threadInfos[:min(3, len(threadInfos))] {
        fmt.Printf("Rank %d: TID=%s, CPU=%.2f%%, Stack:\n%s\n",
            i+1, info.TID, info.CPU, info.Stack)
    }

    // 按内存排序
    sort.Slice(threadInfos, func(i, j int) bool {
        return threadInfos[i].Memory > threadInfos[j].Memory
    })
    fmt.Println("Top 3 threads by Memory usage:")
    for i, info := range threadInfos[:min(3, len(threadInfos))] {
        fmt.Printf("Rank %d: TID=%s, Memory=%s, Stack:\n%s\n",
            i+1, info.TID, FormatMemory(info.Memory), info.Stack)
    }
}
```
