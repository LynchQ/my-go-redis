package config

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/LynchQ/my-go-redis/lib/logger"
)

// ServerProperties 定义全局配置属性
type ServerProperties struct {
	Bind        string   `cfg:"bind"`        // 监听地址
	Port        int      `cfg:"port"`        // 监听端口
	AppendOnly  bool     `cfg:"appendOnly"`  // 是否开启持久化
	MaxClient   int      `cfg:"maxClient"`   // 最大客户端连接数
	RequirePass string   `cfg:"requirepass"` // 密码
	Databases   int      `cfg:"databases"`   // 数据库数
	Peers       []string `cfg:"peers"`       // 集群节点
	Self        string   `cfg:"self"`        // 本节点
}

// Properties 保存全局配置属性
var Properties *ServerProperties

func init() {
	// 默认配置
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: false,
	}
}

func parse(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	// 读取配置文件
	rawMap := make(map[string]string)
	// 创建一个扫描器
	scanner := bufio.NewScanner(src)
	// 逐行扫描
	for scanner.Scan() {
		// 读取一行
		line := scanner.Text()
		// 如果是注释则跳过
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		// 找到分隔符
		// strings.IndexAny从s中查找chars中任意一个Unicode字符的第一个实例的索引，
		// 如果s中没有chars中的任何一个Unicode字符则返回-1
		pivot := strings.IndexAny(line, " ")
		// 找到分隔符
		if pivot > 0 && pivot < len(line)-1 {
			// q: 为什么要判断pivot < len(line)-1
			//a: 因为pivot是分隔符的位置，如果pivot是最后一个字符，那么pivot+1就会越界
			// 分隔符前面的是key，后面的是value
			key := line[0:pivot]
			// strings.Trim 去掉value前后的空格
			value := strings.Trim(line[pivot+1:], " ")
			// 将key转换为小写，然后存入rawMap
			rawMap[strings.ToLower(key)] = value
		}
	}
	// 如果扫描器出错，则打印错误
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	// 通过反射，将rawMap中的值赋值给config
	// t 是config的类型
	t := reflect.TypeOf(config)
	// v 是config的值
	v := reflect.ValueOf(config)
	// n 是config的字段数
	n := t.Elem().NumField()

	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			// 根据fieldVal的类型，将value转换为对应的类型
			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				// 将value转换为int64
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				// 将value转换为小写，然后判断是否等于yes
				boolValue := "yes" == value
				fieldVal.SetBool(boolValue)
			case reflect.Slice:
				// 将value按照逗号分隔，然后存入fieldVal
				if field.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					fieldVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	// 返回config
	return config
}

// SetupConfig 读取配置文件并且初始化配置
func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	Properties = parse(file)
}
