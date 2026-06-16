package db

import (
	"fmt"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB wraps the GORM database connection
type DB struct {
	*gorm.DB
}

// Models

type PinState struct {
	ID        uint   `gorm:"primarykey"`
	NodeID    string `gorm:"index"`
	PinID     string `gorm:"index"`
	State     string // "HIGH", "LOW", or JSON for analog
	Timestamp int64  `gorm:"autoCreateTime"`
}

type ActionLog struct {
	ID          uint   `gorm:"primarykey"`
	NodeID      string `gorm:"index"`
	PinID       string
	Action      string
	Params      string // JSON
	Result      string
	TriggeredBy string // "user", "automation:<id>", "system"
	Timestamp   int64  `gorm:"autoCreateTime"`
}

type Automation struct {
	ID        string `gorm:"primarykey"`
	Name      string
	Enabled   bool
	Config    string // Full JSON representation
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	if dbPath == "" {
		dbPath = "pi-gpio.db"
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "/" {
		// We can't create dirs in this simple version, assume it exists
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(&PinState{}, &ActionLog{}, &Automation{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{db}, nil
}

// LogAction records an action in the database
func (d *DB) LogAction(nodeID, pinID, action, params, result, triggeredBy string) error {
	return d.Create(&ActionLog{
		NodeID:      nodeID,
		PinID:       pinID,
		Action:      action,
		Params:      params,
		Result:      result,
		TriggeredBy: triggeredBy,
	}).Error
}

// GetActionLogs retrieves recent action logs
func (d *DB) GetActionLogs(limit int, nodeID string) ([]ActionLog, error) {
	var logs []ActionLog
	query := d.DB.Order("timestamp DESC").Limit(limit)
	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	err := query.Find(&logs).Error
	return logs, err
}

// SaveAutomation saves an automation rule
func (d *DB) SaveAutomation(automation *Automation) error {
	return d.Save(automation).Error
}

// GetAutomations retrieves all automation rules
func (d *DB) GetAutomations() ([]Automation, error) {
	var automations []Automation
	err := d.Find(&automations).Error
	return automations, err
}

// GetAutomation retrieves a single automation rule
func (d *DB) GetAutomation(id string) (*Automation, error) {
	var automation Automation
	if err := d.First(&automation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &automation, nil
}

// DeleteAutomation removes an automation rule
func (d *DB) DeleteAutomation(id string) error {
	return d.Delete(&Automation{}, "id = ?", id).Error
}

// TrimActionLogs removes old action logs, keeping only the most recent N entries
func (d *DB) TrimActionLogs(maxEntries int) error {
	var count int64
	if err := d.Model(&ActionLog{}).Count(&count).Error; err != nil {
		return err
	}

	if count <= int64(maxEntries) {
		return nil
	}

	// Delete oldest entries
	toDelete := count - int64(maxEntries)
	return d.Where("id IN (?)", 
		d.Model(&ActionLog{}).
			Select("id").
			Order("timestamp ASC").
			Limit(int(toDelete)),
	).Delete(&ActionLog{}).Error
}
