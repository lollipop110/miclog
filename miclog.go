// miclog project miclog.go
package miclog

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type MicLog struct {
	path     string         //文件路径
	name     string         //文件名称
	maxSize  int            //最大尺寸(kb)
	saveDay  int            //最大存储时间(day)
	chansize int            //缓冲池大小,默认100
	print    bool           //是否在控制台打印显示
	cache    chan logStruct //缓存
}

type logStruct struct {
	Time    time.Time
	Type    string
	Message string
}

var _miclog MicLog

/*********************************************************************
功  能:初始化
参  数:无
返回值:无
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func init() {
	_miclog.path = ""
	_miclog.name = "Miclog"
	_miclog.maxSize = 102400
	_miclog.saveDay = 30
	_miclog.chansize = 100
	_miclog.cache = make(chan logStruct, _miclog.chansize)

	go _miclog.run()
}

/*********************************************************************
功  能:关闭
参  数:无
返回值:无
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func (log *MicLog) close() {
	close(log.cache)
}

/*********************************************************************
功  能:用户配置信息
参  数:
	   logpath string:日志文件存储路径
	   logname string:日志文件名称
	   maxsize int:日志文件最大字节数
	   saveday int:保存时间(天)
返回值:无
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func Config(logpath, logname string, maxsize, saveday int) {
	_miclog.path = logpath
	_miclog.name = logname
	_miclog.maxSize = maxsize
	_miclog.saveDay = saveday
}

/*********************************************************************
功  能:写入 Info 类型的日志
参  数:
	   log string:日志信息
	   args interface{}:日志信息中的可选参数
返回值:无
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func Info(log string, args ...interface{}) {
	lg := newlog(5, log, args...)
	_miclog.cache <- lg
}

/*********************************************************************
功  能:新的日志信息
参  数:
	   logtype int:日志类型.
			0:EMER 系统级紧急，比如磁盘出错，内存异常，网络不可用等	红色底
			1:ALRT 系统级警告，比如数据库访问异常，配置文件出错等	紫色
			2:CRIT 系统级危险，比如权限出错，访问异常等	蓝色
			3:EROR 用户级错误	红色
			4:WARN 用户级警告	黄色
			5:INFO 用户级重要	天蓝色
			6:DEBG 用户级调试	绿色
			7:TRAC 用户级基本输出	绿色用户级基本输出	绿色
	   log string:日志信息
	   args interface{}:日志信息中的可选参数
返回值:无
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func newlog(logtype int, log string, args ...interface{}) logStruct {
	var lg logStruct
	lg.Message = fmt.Sprintf(log, args...)
	lg.Time = time.Now()
	switch logtype {
	case 0: //系统级紧急，比如磁盘出错，内存异常，网络不可用等	红色底
		lg.Type = "EMER"
	case 1: //系统级警告，比如数据库访问异常，配置文件出错等	紫色
		lg.Type = "ALRT"
	case 2: //系统级危险，比如权限出错，访问异常等	蓝色
		lg.Type = "CRIT"
	case 3: //用户级错误	红色
		lg.Type = "EROR"
	case 4: //用户级警告	黄色
		lg.Type = "WARN"
	case 5: //用户级重要	天蓝色
		lg.Type = "INFO"
	case 6: //用户级调试	绿色
		lg.Type = "DEBG"
	case 7: //用户级基本输出	绿色用户级基本输出	绿色
		lg.Type = "TRAC"
	}
	if _miclog.print {
		fmt.Printf("%s [%s] %s\r", lg.Time.Format("2006-01-02 15:04:05.000"), lg.Type, lg.Message)
	}
	return lg
}

/*********************************************************************
函数名:Run
参  数:启动线程
返回值:无
创建时间:2018年12月20日
修订信息:
*********************************************************************/
func (log *MicLog) run() {
	defer log.close()
	go log.checklog()
	for {
		time.Sleep(1 * time.Second)
		if len(log.cache) > 0 {
			log.WriteLog()
		}
	}
}

/*********************************************************************
功  能:检查日志文件是否超限
参  数:无
返回值:无
说  明:用作go线程,每小时执行一次
创建时间:2020年10月26日
修订信息:
*********************************************************************/
func (log *MicLog) checklog() {
	time.Sleep(1 * time.Minute)
	for {
		log.checkLogFiles()
		time.Sleep(60 * time.Minute)
	}
}

/*********************************************************************
函数名:写日志信息进入文件
参  数:content string:需要记录的信息
	isprint bool:是否在控制台打印显示
返回值:无
创建时间:2018年12月20日
修订信息:
*********************************************************************/
func (log *MicLog) WriteLog() {
	if err := os.MkdirAll(log.path, os.ModePerm); err != nil {
		fmt.Println("Creat dir fault[创建日志文件夹错误]", err.Error())
	}
	name := fmt.Sprintf("%s/%s_%s.log", log.path, log.name, time.Now().Format("2006-01-02"))
	fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("打开文件错误:", err.Error())
		return
	}
	defer fileObj.Close()
	writeObj := bufio.NewWriterSize(fileObj, 4096)
	for lg := range log.cache {
		msg := fmt.Sprintln(lg.Time.Format("2006-01-02 15:04:05.000"), "[", lg.Type, "]", lg.Message)
		fmt.Print(msg)
		//buf = append(buf, []byte(msg)...)
		_, err := writeObj.WriteString(msg)
		if err != nil {
			fmt.Printf("写日志错误[%s]", err.Error())
		}
		if err := writeObj.Flush(); err != nil {
			fmt.Printf("写日志错误[%s]", err.Error())
		}
	}
}

/***************************************************************
功  能:检查日志文件是否存在尺寸过大或者存储时间超时的文件,有则删除之
参  数:
返回值:无
创建时间:2019年2月1日
修订信息:
***************************************************************/
func (log *MicLog) checkLogFiles() {
	dir_list, e := ioutil.ReadDir(log.path)
	if e != nil {
		fmt.Println("read dir error")
		return
	}
	delflag := 0
	for _, v := range dir_list {
		delflag = 0
		//fmt.Println(i, "=", v.Name(), v.ModTime(), v.Size())
		if log.maxSize > 0 && v.Size() > int64(log.maxSize)*1024 { //文件尺寸超限
			delflag = 1
		} else {
			l := len(v.Name())
			if log.saveDay > 0 && l > 14 {
				fileDate := string([]byte(v.Name())[l-14 : l-4])
				t, e := time.ParseInLocation("2006-01-02", fileDate, time.Local)
				if e == nil {
					//fmt.Println("所有:", v.Name(), t)
					if t.Before(time.Now().Add(-24 * time.Hour * time.Duration(log.saveDay))) {
						delflag = 2
					}
				}
			}
		}
		if delflag > 0 {
			if strings.Contains(v.Name(), ".log") { //必须是log文件才可以删除
				file := fmt.Sprintf("%s/%s", log.path, v.Name()) //文件路径
				msg := ""
				if delflag == 1 {
					msg = fmt.Sprintf("日志文件[%s]超过[%d]KB,删除之！", file, log.maxSize)
				} else {
					msg = fmt.Sprintf("日志文件[%s]超过[%d]天,删除之！", file, log.saveDay)
				}
				//log.WriteLog(msg, true) //记录消息
				log.cache <- newlog(0, msg)
				err := os.Remove(file) //删除文件
				if err != nil {
					//如果删除失败则输出 file remove Error!
					msg = fmt.Sprintf("删除日志文件[%s]失败!错误信息:[%s]", file, err)
					//log.WriteLog(msg, true) //记录消息
					log.cache <- newlog(0, msg)
				}
			}
		}
	}
}
