package speed

import (
	"encoding/json"
	"log"

	"github.com/gbsto/daisy/db"
)

func GetSpeedChartData() string {
	type chartData struct {
		Timestamps []int64   `json:"timestamps"`
		Download   []float64 `json:"download"`
		Upload     []float64 `json:"upload"`
	}

	data := chartData{
		Timestamps: make([]int64, 0, 30),
		Download:   make([]float64, 0, 30),
		Upload:     make([]float64, 0, 30),
	}
	rows, err := db.Conn.Query("SELECT Created, DownloadSpeed, UploadSpeed FROM speedtest ORDER BY Created DESC LIMIT 30")
	if err != nil {
		log.Println(err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var timestamp int64
		var download float64
		var upload float64
		if err := rows.Scan(&timestamp, &download, &upload); err != nil {
			log.Println(err)
			return ""
		}
		data.Timestamps = append(data.Timestamps, timestamp)
		data.Download = append(data.Download, download)
		data.Upload = append(data.Upload, upload)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return ""
	}

	for left, right := 0, len(data.Timestamps)-1; left < right; left, right = left+1, right-1 {
		data.Timestamps[left], data.Timestamps[right] = data.Timestamps[right], data.Timestamps[left]
		data.Download[left], data.Download[right] = data.Download[right], data.Download[left]
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(encoded)
}
