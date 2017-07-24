package main

import (
	_ "database/sql"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Datastore struct {
	db       *sqlx.DB
	filename string
}

type Device struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Color        string    `json:"color" db:"color"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	Disabled     bool      `json:"disabled" db:"disabled"`
	Error        string    `json:"error" db:"error"`
	LatestMetric Metric    `json:"latest_metrics" db:"latest"`
	Created      time.Time `json:"created" db:"created"`
	Updated      time.Time `json:"updated" db:"updated"`
}

type Metric struct {
	ID          int       `json:"-" db:"id"`
	DeviceID    string    `json:"-" db:"device_id"`
	Power       int       `json:"power" db:"power"`
	Battery     int       `json:"battery" db:"battery"`
	Temperature int       `json:"temperature" db:"temperature"`
	Gravity     float64   `json:"gravity" db:"gravity"`
	Created     time.Time `json:"created" db:"created"`
}

func NewDatastore(filename string) *Datastore {
	db, err := sqlx.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS device (
	id VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255),
	color VARCHAR(50),
	endpoint TEXT,
	disabled BOOLEAN,
	error TEXT,
	created TIMESTAMP,
	updated TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS metric (
	id INTEGER PRIMARY KEY,
	device_id VARCHAR(255) NOT NULL,
	power INTEGER,
	battery INTEGER,
	temperature INTEGER,
	gravity REAL,
	created TIMESTAMP
	);
	`
	db.MustExec(query)

	return &Datastore{
		db:       db,
		filename: filename,
	}
}

func (d *Datastore) Close() {
	d.db.Close()
}

func (d *Datastore) GetDevices() ([]Device, error) {
	devices := []Device{}
	if err := d.db.Select(&devices, "SELECT * FROM device"); err != nil {
		return devices, err
	}
	return devices, nil
}

func (d *Datastore) GetDevicesWithMetrics() ([]Device, error) {
	query := `
	SELECT d.*,
	m1.power AS "latest.power",
	m1.battery AS "latest.battery",
	m1.temperature AS "latest.temperature",
	m1.gravity AS "latest.gravity",
	m1.created AS "latest.created"
	FROM device d
  	JOIN metric m1 ON (d.id = m1.device_id)
  	LEFT OUTER JOIN metric m2 ON (d.id = m2.device_id AND
    (m1.created < m2.created OR m1.created = m2.created AND m1.id < m2.id))
    WHERE m2.id IS NULL;
	`

	devices := []Device{}
	rows, err := d.db.Queryx(query)
	for rows.Next() {
		device := Device{}
		if err := rows.StructScan(&device); err != nil {
			return devices, err
		}

		devices = append(devices, device)
	}
	return devices, err
}

func (d *Datastore) GetDevice(id string) (Device, error) {
	device := Device{}
	err := d.db.Get(&device, "SELECT * FROM device WHERE id=$1", id)
	return device, err
}

func (d *Datastore) CreateOrUpdateDevice(device Device) error {
	statement, _ := d.db.Prepare("INSERT OR IGNORE INTO device VALUES (?,?,?,?,?,?,?,?)")
	_, err := statement.Exec(
		device.ID,
		device.Name,
		device.Color,
		device.Endpoint,
		device.Disabled,
		device.Error,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return d.UpdateDevice(device)
}

func (d *Datastore) UpdateDevice(device Device) error {
	statement, _ := d.db.Prepare("UPDATE device SET name=?, endpoint=?, disabled=?, updated=? WHERE id=?")
	_, err := statement.Exec(
		device.Name,
		device.Endpoint,
		device.Disabled,
		time.Now(),
		device.ID,
	)
	return err
}

func (d *Datastore) DeleteDevice(id string) error {
	_, err := d.db.Exec("DELETE FROM device WHERE id=$1", id)
	return err
}

func (d *Datastore) SetDeviceError(id string, errorMsg string) error {
	_, err := d.db.Exec("UPDATE device SET error=$1 WHERE id=$2", errorMsg, id)
	return err
}

func (d *Datastore) CreateMetric(metric Metric) error {
	statement, _ := d.db.Prepare("INSERT INTO metric (device_id, power, battery, temperature, gravity, created) VALUES (?,?,?,?,?,?)")
	_, err := statement.Exec(
		metric.DeviceID,
		metric.Power,
		metric.Battery,
		metric.Temperature,
		metric.Gravity,
		time.Now(),
	)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("UPDATE device SET updated=$1 WHERE id=$2", time.Now(), metric.DeviceID)
	return err
}

func (d *Datastore) GetDeviceMetrics(id string, limit string) ([]Metric, error) {
	metrics := []Metric{}
	err := d.db.Select(&metrics, "SELECT * FROM metric WHERE device_id=$1 ORDER BY created ASC LIMIT $2", id, limit)
	return metrics, err
}

func (d *Datastore) GetDeviceLatestMetrics(id string) (Metric, error) {
	metric := Metric{}
	err := d.db.Get(&metric, "SELECT * FROM metric WHERE device_id=$1 ORDER BY created ASC LIMIT 1", id)
	return metric, err
}
