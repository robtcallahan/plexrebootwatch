package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Config struct {
	Phone   string `json:"phone"`
	LogDir  string `json:"log_dir"`
	LogFile string `json:"log_file"`
}

func main() {
	config, err := loadConfig("/usr/local/etc/plexrebootwatch.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	boot, err := getBootTime()
	if err != nil {
		log.Fatalf("Failed to get boot time: %v", err)
	}

	now := time.Now().Format(time.RFC1123)
	uptime := formatUptime(boot)
	body := fmt.Sprintf("System rebooted at: %s\nCurrent uptime: %s\nChecked at: %s",
		boot.Format(time.RFC1123), uptime, now)

	if rebootedRecently(boot) {
		if err := sendText(config.Phone, body); err != nil {
			log.Fatalf("Failed to send email: %v", err)
		}
		fmt.Println("Reboot email sent.")
	} else {
		fmt.Println("No recent reboot.")
	}
}

func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getBootTime() (time.Time, error) {
	// Read kern.boottime from sysctl
	raw, err := unix.SysctlRaw("kern.boottime")
	if err != nil {
		return time.Time{}, err
	}
	// Convert raw bytes into Timeval (tv_sec, tv_usec)
	if len(raw) < 16 {
		return time.Time{}, fmt.Errorf("unexpected raw length: %d", len(raw))
	}
	tv := *(*unix.Timeval)(unsafe.Pointer(&raw[0]))
	return time.Unix(int64(tv.Sec), int64(tv.Usec)*1000), nil
}

func rebootedRecently(boot time.Time) bool {
	return time.Since(boot) <= 15*time.Minute
}

func formatUptime(boot time.Time) string {
	uptime := time.Since(boot)
	hours := int(uptime.Hours())
	d := hours / 24
	h := hours - (d * 24)
	m := int(uptime.Minutes()) % 60
	s := int(uptime.Seconds()) % 60
	return fmt.Sprintf("%d days %02d:%02d:%02d", d, h, m, s)
}

func sendText(phone, message string) error {
	const templateString = `
			tell application "Messages"
				set targetService to 1st service whose service type = iMessage
				set targetBuddy to buddy "{{.Number}}" of targetService
				send "{{.Message}}" to targetBuddy
			end tell
			`

	// Create a new template and parse the template string
	tmpl, err := template.New("greeting").Parse(templateString)
	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{
		"Number":  phone,
		"Message": message,
	}

	// Execute the template and write the output to standard output
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	cmd := exec.Command("osascript", "-e", buf.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if len(output) != 0 {
		fmt.Println("Output:", output)
	}
	return nil
}
