# atmo
atmo is a tinygo package wrapping a bme280 (thermometer, barometer, pressure, altimeter) with some new useful methods

### Intended use
- Connect a BME280 chip to the I2C bus of a microcontroller
- Set up a goroutine which reads the sensor & updates the state
- Employ a `trigger.Dispatcher` which listens for triggers intended for the sensor
  - The dispatcher calls `Execute(trigger.Trigger)`
- The `trigger.Trigger` has a channel on which to send mutated trigger, 


### `New(*bme280, name)`
- pass a pre-configured bme280 and a name

### Options for using 

### `Execute(trigger.Trigger)`
synchronously returns a trigger filled with a `Message` containing the requested information
#### Valid actions
- All - returns Temperature, Humidity, Pressure, Altitude, each including the time of the last reading
- Temp - returns Temperature & the time of the last reading
- Humi - returns Humidity and the time of the last reading
- Baro - returns Pressure and the time of the last reading
- Alti - returns Altitude and the time of the last reading