package atmo

import (
	"strconv"
	"strings"
	"time"

	"github.com/eyelight/trigger"
	"tinygo.org/x/drivers/bme280"
)

type atmo struct {
	bme  *bme280.Device
	name string
	temp state // value is temperature in milli-Celsius (°C * 1000; c = val/1000)
	baro state // value is pressure in milliPascals (mPa)
	humi state // value is humitidy as hundredths of %
	alti state // value is meters of elevation
}

type state struct {
	name  string
	value int32
	since time.Time
}

type Atmo interface {
	Name() string
	Execute(t trigger.Trigger)
	StateString() string
	TempString() string
	BaroString() string
	HumiString() string
	AltiString() string
	Temp() (int32, error)
	Baro() (int32, error)
	Humi() (int32, error)
	Alti() (int32, error)
	Update() error
	ResetAll()
	Connected() bool
}

// New returns a new Atmo object with zeroed-out states; pass a configured bme280.Device and a name
func New(b *bme280.Device, n string) Atmo {
	t := time.Now()
	return &atmo{
		bme:  b,
		name: n,
		temp: state{name: "Temperature", value: 0, since: t},
		baro: state{name: "Barometer", value: 0, since: t},
		humi: state{name: "Pressure", value: 0, since: t},
		alti: state{name: "Altitude", value: 0, since: t},
	}
}

func (s *state) reset() {
	s.value = 0
	s.since = time.Now()
}

func (s *state) string(conv, units string) string {
	ss := strings.Builder{}
	ss.Grow(512)
	ss.WriteByte(byte(9)) // tab
	ss.WriteString(s.name)
	ss.WriteByte(byte(9)) // tab
	ss.WriteString(conv)  // spit out the passed-in converted string
	ss.WriteString(units) // spit out the units
	ss.WriteString(" (since ")
	ss.WriteString(s.since.Local().String())
	ss.WriteString(")")
	ss.WriteByte(byte(10)) // newline
	return ss.String()
}

// Connected returns a bool representing whether or not the bme280.Device is connected
func (a *atmo) Connected() bool {
	return a.bme.Connected()
}

func (a *atmo) ResetAll() {
	a.temp.reset()
	a.baro.reset()
	a.humi.reset()
	a.alti.reset()
}

// Update returns an error or nil after calling read/update methods
func (a *atmo) Update() error {
	_, err := a.Temp()
	if err != nil {
		return err
	}
	_, err = a.Baro()
	if err != nil {
		return err
	}
	_, err = a.Humi()
	if err != nil {
		return err
	}
	_, err = a.Alti()
	if err != nil {
		return err
	}
	return nil
}

// Temp returns an int32 in celsius milli degrees & an error, updating internal state if no error
func (a *atmo) Temp() (int32, error) {
	t, err := a.bme.ReadTemperature()
	if err != nil {
		return (-420), err
	}
	a.temp.value = t
	return a.temp.value, nil
}

func (a *atmo) TempString() string {
	return a.temp.string(strconv.FormatFloat(ctof(a.Celsius()), 'f', 2, 64), "°F")
}

func (a *atmo) BaroString() string {
	return a.baro.string(strconv.FormatFloat(mpatoin(a.baro.value), 'f', 2, 64), "\" Hg")
}

func (a *atmo) HumiString() string {
	return a.humi.string(strconv.FormatFloat(toPct(a.humi.value), 'f', 2, 64), "%")
}

func (a *atmo) AltiString() string {
	return a.alti.string(strconv.FormatFloat(a.FeetElevation(), 'f', 2, 64), "ft")
}

// Baro returns the barometric pressure in millipascals (mPa) & and error, updating internal state if no error
func (a *atmo) Baro() (int32, error) {
	b, err := a.bme.ReadPressure()
	if err != nil {
		return (-420), err
	}
	a.baro.value = b
	return a.baro.value, nil
}

// Humi returns the humidity in hundredths of percent; convert to a float somewhere else
func (a *atmo) Humi() (int32, error) {
	h, err := a.bme.ReadHumidity()
	if err != nil {
		return (-420), err
	}
	a.humi.value = h
	return a.humi.value, nil
}

// Alti returns the altitude in meters, by wrapping a call to (*bme280.Device).ReadAltitude()
func (a *atmo) Alti() (int32, error) {
	alt, err := a.bme.ReadAltitude()
	if err != nil {
		return (-420), err
	}
	a.alti.value = alt
	return a.alti.value, nil
}

// State is a Statist interface method, returning the current state and the time.Time 'since' the state was
func (a *atmo) State() (interface{}, time.Time) {
	return nil, time.Now()
}

// StateString is a Statist interface method, returning a formatted string of current state
func (a *atmo) StateString() string {
	ss := strings.Builder{}
	ss.Grow(1024)
	// report top level device name etc
	ss.WriteString(a.name)
	ss.WriteByte(byte(10)) // newline

	// report temp with converted temp & unit
	ss.WriteString(a.temp.string(strconv.FormatFloat(ctof(a.Celsius()), 'f', 2, 64), "°F"))
	// report baro with converted pressure & unit
	ss.WriteString(a.baro.string(strconv.FormatFloat(mpatoin(a.baro.value), 'f', 2, 64), "\" Hg"))
	// report humi with converted humidity & unit
	ss.WriteString(a.humi.string(strconv.FormatFloat(toPct(a.humi.value), 'f', 2, 64), "%"))
	// report alti with converted elevation & unit
	ss.WriteString(a.alti.string(strconv.FormatFloat(a.FeetElevation(), 'f', 2, 64), "ft"))
	return ss.String()
}

// Name is a Statist interface method, returning the internal name
func (a *atmo) Name() string {
	return a.name
}

func (a *atmo) Execute(t trigger.Trigger) {
	println("atmo Execute()...")
	t.Error = false
	if t.Target != a.name {
		t.Error = true
		t.Message = string("error - " + a.name + " received a trigger intended for " + t.Target)
		t.ReportCh <- t
		return
	}
	err := a.Update()
	if err != nil {
		t.Error = true
		t.Message = string("error - " + a.name + " couldn't get sensor update - " + err.Error())
		t.ReportCh <- t
		return
	}
	switch t.Action {
	case "All", "all", "":
		t.Message = string(a.name + " Report: " + a.StateString())
	case "Temp", "temp", "Temperature", "temperature":
		t.Message = string(a.name + " Temp: " + a.TempString())
	case "Hum", "hum", "Humidity", "humidity":
		t.Message = string(a.name + " Humi: " + a.HumiString())
	case "Pres", "pres", "Pressure", "pressure", "Baro", "baro", "Barometer", "barometer":
		t.Message = string(a.name + "Baro: " + a.BaroString())
	case "Alti", "alti", "Altitude", "altitude":
		t.Message = string(a.name + "Alt: " + a.AltiString())
	}
	t.ReportCh <- t
	return
}

// Celsius converts the latest stored milli-Celsius to Celsius & returns it
func (a *atmo) Celsius() float64 {
	return float64(a.temp.value / 1000)
}

// FeetElevation converts the latest stored meters to feet & returns it
func (a *atmo) FeetElevation() float64 {
	return mtof(a.alti.value)
}

// Farenheit converts the latest stored milli-Celsius to Farenheit & returns it
func (a *atmo) Farenheit() float64 {
	return ctof(a.Celsius())
}

// ctof converts celsius to farenheit
func ctof(c float64) float64 {
	return (c * 1.8) + 32
}

// mtof converts meters to feet
func mtof(m int32) float64 {
	// feet = m * 3.2808
	return float64(m) * 3.2808
}

// toPct returns percentage from a passed-in hundredths of a percent
func toPct(h int32) float64 {
	return float64(h) / 100
}

// mpatoin converts mPa to inches of mercury
func mpatoin(m int32) float64 {
	return float64(m) / 3386389
}
