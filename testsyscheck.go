package main

import (
	"database/sql" //sql pack
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux" //cpu_check package
	_ "github.com/go-sql-driver/mysql"
	memory "github.com/pbnjay/memory" //Memory check package
	"log"
	"strconv" //typecasting package
	"syscall" //disk check package
	"time"
)

//Disk sturcture
type DiskStatus struct {
	All  uint64 `json:"All"`
	Used uint64 `json:"Used"`
	Free uint64 `json:"Free"`
}

//convert Byte unit
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func DiskUsage() (arr_Disk []string) {
	//check Disk
	fs := syscall.Statfs_t{}
	//err := syscall.Statfs(path, &fs)
	err := syscall.Statfs("/", &fs)
	if err != nil {
		return
	}
	//DiskStatus struct
	var disk DiskStatus

	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free

	//Disk usage calculata & convert Btye to Giga Byte
	disk_all := float64(disk.All) / float64(GB)
	disk_free := float64(disk.Free) / float64(GB)
	disk_used := float64(disk.Used) / float64(GB)
	disk_percent := float64(disk.Used) / float64(disk.All) * 100

	//Disk usage convert float to string
	Sdisk_all := strconv.FormatFloat(disk_all, 'E', -1, 64)
	Sdisk_free := strconv.FormatFloat(disk_free, 'E', -1, 64)
	Sdisk_used := strconv.FormatFloat(disk_used, 'E', -1, 64)
	Sdisk_percent := strconv.FormatFloat(disk_percent, 'E', -1, 64)

	// DB Insert into parameter string
	arr_Disk = []string{"DISK", "DISK_ALL_GB", "DISK_FREE_GB", "DISK_USED_GB", "DISK_USED_PERCENT",
		Sdisk_all, Sdisk_free, Sdisk_used, Sdisk_percent}

	return
}

//arr_Cpu = Query
func CpuUsage() (arr_Cpu []string) {

	//CPU usage file read & error check
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Fatal("stat read fail")
	}

	// User_uage : User 사용 퍼센트, System_usage : System 사용 퍼센트
	var CPU_all uint64 = stat.CPUStatAll.User + stat.CPUStatAll.Nice + stat.CPUStatAll.System + stat.CPUStatAll.Idle
	var User_usage float64 = float64(stat.CPUStatAll.User) * 100 / float64(CPU_all)
	var System_usage float64 = float64(stat.CPUStatAll.System) * 100 / float64(CPU_all)

	//Disk usage convert float to string
	SUser_usage := strconv.FormatFloat(User_usage, 'E', -1, 64)
	SSystem_usage := strconv.FormatFloat(System_usage, 'E', -1, 64)

	// DB Insert into parameter string
	arr_Cpu = []string{"CPU", "USER_MODE_PERCENT", "SYSTEM_MODE_PERCENT", SUser_usage, SSystem_usage}

	return
}

func MemoryUsage() (arr_Memory []string) {

	//Memory_ B -> GB
	Memory_total_GB := float64(memory.TotalMemory()) / float64(GB)
	Memory_free_GB := float64(memory.FreeMemory()) / float64(GB)
	Memory_used_GB := float64(memory.TotalMemory()-memory.FreeMemory()) / float64(GB)
	Memory_percent := float64(memory.TotalMemory()-memory.FreeMemory()) / float64(memory.TotalMemory()) * 100

	// float to String
	Str_Memory_all := strconv.FormatFloat(Memory_total_GB, 'E', -1, 64)
	Str_Memory_free := strconv.FormatFloat(Memory_free_GB, 'E', -1, 64)
	Str_Memory_used := strconv.FormatFloat(Memory_used_GB, 'E', -1, 64)
	Str_Memory_percent := strconv.FormatFloat(Memory_percent, 'E', -1, 64)

	//Make array for Insert DB
	arr_Memory = []string{"MEMORY", "MEMORY_ALL_GB", "MEMORY_USED_GB", "MEMORY_FREE_GB", "MEMORY_USED_PERCENT",
		Str_Memory_all, Str_Memory_free, Str_Memory_used, Str_Memory_percent}

	return

}

//INSERT INTO USER(column1, column2, ...) VALUES (value1, value2, ...)
func DbInsert(Query_input []string) {

	//1p : drive name	2p : connector info
	db, err := sql.Open("mysql", "test:test@test(111.1.1.1)")
	if err != nil {
		log.Fatal(err)
	}
	//Before query, use it (reset == ;)
	//defer function used before current function end
	defer db.Close()

	//Make "Insert Query"
	std_len := (len(Query_input) - 1) / 2
	var Query string = "INSERT INTO " + Query_input[0] + "("

	for i := 1; i < std_len+1; i++ {
		if i == 1 {
			Query += Query_input[i]
			continue
		}
		Query = Query + "," + Query_input[i]
	}
	Query = Query + ") VALUES ("
	for i := std_len + 1; i < len(Query_input); i++ {
		if i == std_len+1 {
			Query += Query_input[i]
			continue
		}
		Query = Query + "," + Query_input[i]
	}

	Query = Query + ")"

	//Querying & error check
	result, err := db.Exec(Query)
	if err != nil {
		log.Fatal(err)
	}

	//Server response check
	n, err := result.RowsAffected()
	if n == 1 {
		switch Query_input[0] {
		case "DISK":
			fmt.Println("DISK row inserted.")
		case "CPU":
			fmt.Println("CPU row inserted.")
		case "MEMORY":
			fmt.Println("MEMORY row inserted.")
		}
	}
}

func main() {

	//infinite loop
	for {
		//Disk usage insert
		DbInsert(DiskUsage())

		//CPU usage insert
		DbInsert(CpuUsage())
		
		//Memory usage insert
		DbInsert(MemoryUsage())

		time.Sleep(time.Minute)
	}

}
