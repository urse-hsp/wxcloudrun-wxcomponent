package main

// 临时只读探针：验证 authorizers.updatetime 的秒级截断是否会导致
// ClearAuthorizerRecordsBefore(now) 把刚写入的记录立刻删掉。
// 全程使用 TEMPORARY TABLE，仅在自身连接内存在，断开即消失，不触碰真实数据。
import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	user := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	addr := os.Getenv("MYSQL_ADDRESS")
	name := os.Getenv("MYSQL_DATABASE")
	if name == "" {
		name = "wxcomponent"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true&loc=Local", user, pwd, addr, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("open failed:", err)
		return
	}
	defer db.Close()
	// 临时表与连接绑定，必须全程复用同一个连接
	db.SetMaxOpenConns(1)

	fmt.Println("========== authorizers.updatetime 列定义 ==========")
	rows, err := db.Query(`SELECT COLUMN_NAME, COLUMN_TYPE, DATETIME_PRECISION, EXTRA
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'authorizers' AND COLUMN_NAME IN ('updatetime','authtime','createtime')`)
	if err != nil {
		fmt.Println("查询列定义失败:", err)
	} else {
		for rows.Next() {
			var col, ctype, extra string
			var prec sql.NullInt64
			rows.Scan(&col, &ctype, &prec, &extra)
			fmt.Printf("  %-12s type=%-12s 小数位精度=%v  extra=%s\n", col, ctype, prec.Int64, extra)
		}
		rows.Close()
	}

	fmt.Println("\n========== 当前 authorizers 真实记录数 ==========")
	var cnt int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM authorizers`).Scan(&cnt); err != nil {
		fmt.Println("查询失败:", err)
	} else {
		fmt.Println("  记录数:", cnt)
	}

	fmt.Println("\n========== 复现 pull 的写入+清理时序（临时表，零风险）==========")
	stmts := []string{
		`DROP TEMPORARY TABLE IF EXISTS t_probe`,
		`CREATE TEMPORARY TABLE t_probe LIKE authorizers`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			fmt.Println("准备临时表失败:", s, err)
			return
		}
	}

	// 模拟 pullAuthorizerListHandler: 先记录 now，再请求微信(耗时)，最后写入
	now := time.Now()
	fmt.Printf("  [T0] Go 捕获 now           = %s\n", now.Format("2006-01-02 15:04:05.000000000"))
	time.Sleep(600 * time.Millisecond) // 模拟调用微信接口的耗时

	if _, err := db.Exec(`INSERT INTO t_probe (appid, nickname, authtime) VALUES ('probe_appid', 'probe', NOW())`); err != nil {
		fmt.Println("写入临时表失败:", err)
		return
	}

	var stored time.Time
	var dbNow6 time.Time
	if err := db.QueryRow(`SELECT updatetime, NOW(6) FROM t_probe WHERE appid='probe_appid'`).Scan(&stored, &dbNow6); err != nil {
		fmt.Println("读取临时表失败:", err)
		return
	}
	fmt.Printf("  [T1] 写入后库中 updatetime  = %s   (注意小数部分)\n", stored.Format("2006-01-02 15:04:05.000000000"))

	wouldDelete := stored.Before(now)
	fmt.Printf("  清理条件 updatetime < now  =>  %v < %v  =>  %v\n",
		stored.Format("15:04:05.000000"), now.Format("15:04:05.000000"), wouldDelete)

	if wouldDelete {
		fmt.Println("  >>> 结论：刚写入的记录会被 ClearAuthorizerRecordsBefore(now) 立刻删除")
	} else {
		fmt.Println("  >>> 结论：记录不会被删除（本次时序未命中）")
	}

	// 实际执行一次删除，验证行为
	res, err := db.Exec(`DELETE FROM t_probe WHERE updatetime < ?`, now)
	if err != nil {
		fmt.Println("删除验证失败:", err)
	} else {
		affected, _ := res.RowsAffected()
		var left int64
		db.QueryRow(`SELECT COUNT(*) FROM t_probe`).Scan(&left)
		fmt.Printf("  实际执行 DELETE 后：影响行数=%d，临时表剩余=%d\n", affected, left)
	}

	fmt.Println("\n========== 修复方案验证：now 向下取整到秒并回退 1 秒 ==========")
	fixedNow := now.Truncate(time.Second).Add(-time.Second)
	fmt.Printf("  修正后 now = %s\n", fixedNow.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("  清理条件 updatetime < fixedNow => %v < %v => %v（应为 false，记录得以保留）\n",
		stored.Format("15:04:05.000000"), fixedNow.Format("15:04:05.000000"), stored.Before(fixedNow))

	db.Exec(`DROP TEMPORARY TABLE IF EXISTS t_probe`)
	fmt.Println("\n临时表已删除，未改动任何真实数据。")
}
