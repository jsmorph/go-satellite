package satellite

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Constants
const TWOPI float64 = math.Pi * 2.0
const DEG2RAD float64 = math.Pi / 180.0
const RAD2DEG float64 = 180.0 / math.Pi
const XPDOTP float64 = 1440.0 / (2.0 * math.Pi)

// Holds latitude and Longitude in either degrees or radians
type LatLong struct {
	Latitude, Longitude float64
}

// Holds X, Y, Z position
type Vector3 struct {
	X, Y, Z float64
}

// Holds an azimuth, elevation and range
type LookAngles struct {
	Az, El, Rg float64
}

type Elements struct {
	MeanMotion   float64
	Eccentricity float64
	Inclination  float64
}

// Parses a two line element dataset into a Satellite struct
func ParseTLE(line1, line2, gravconst string) (sat *Satellite, err error) {
	if len(line1) < 61 {
		return nil, fmt.Errorf("Line 1 '%s' too short", line1)
	}

	if len(line2) < 63 {
		return nil, fmt.Errorf("Line 2 '%s' too short", line2)
	}

	sat = &Satellite{}

	sat.Line1 = line1
	sat.Line2 = line2

	sat.Error = 0
	sat.whichconst = getGravConst(gravconst)

	// LINE 1 BEGIN
	sat.satnum, err = parseInt(strings.TrimSpace(line1[2:7]))
	if err != nil {
		return nil, err
	}
	sat.epochyr, err = parseInt(line1[18:20])
	if err != nil {
		return nil, err
	}
	sat.epochdays, err = parseFloat(line1[20:32])
	if err != nil {
		return nil, err
	}

	// These three can be negative / positive
	sat.ndot, err = parseFloat(strings.Replace(line1[33:43], " ", "", 2))
	if err != nil {
		return nil, err
	}

	sat.nddot, err = parseFloat(strings.Replace(line1[44:45]+"."+line1[45:50]+"e"+line1[50:52], " ", "", 2))
	if err != nil {
		return nil, err
	}

	sat.bstar, err = parseFloat(strings.Replace(line1[53:54]+"."+line1[54:59]+"e"+line1[59:61], " ", "", 2))
	if err != nil {
		return nil, err
	}
	// LINE 1 END

	// LINE 2 BEGIN
	sat.inclo, err = parseFloat(strings.Replace(line2[8:16], " ", "", 2))
	if err != nil {
		return nil, err
	}
	sat.nodeo, err = parseFloat(strings.Replace(line2[17:25], " ", "", 2))
	if err != nil {
		return nil, err
	}
	sat.ecco, err = parseFloat("." + line2[26:33])
	if err != nil {
		return nil, err
	}
	sat.argpo, err = parseFloat(strings.Replace(line2[34:42], " ", "", 2))
	if err != nil {
		return nil, err
	}
	sat.mo, err = parseFloat(strings.Replace(line2[43:51], " ", "", 2))
	if err != nil {
		return nil, err
	}
	sat.no, err = parseFloat(strings.Replace(line2[52:63], " ", "", 2))
	if err != nil {
		return nil, err
	}

	// LINE 2 END

	sat.elements = &Elements{
		MeanMotion:   sat.no,
		Eccentricity: sat.ecco,
		Inclination:  sat.inclo,
	}
	return
}

// Converts a two line element data set into a Satellite struct and runs sgp4init
func TLEToSat(line1, line2 string, gravconst string) (*Satellite, error) {
	//sat := Satellite{Line1: line1, Line2: line2}
	sat, err := ParseTLE(line1, line2, gravconst)
	if err != nil {
		return nil, err
	}

	opsmode := "i"

	sat.no = sat.no / XPDOTP
	sat.ndot = sat.ndot / (XPDOTP * 1440.0)
	sat.nddot = sat.nddot / (XPDOTP * 1440.0 * 1440)

	sat.inclo = sat.inclo * DEG2RAD
	sat.nodeo = sat.nodeo * DEG2RAD
	sat.argpo = sat.argpo * DEG2RAD
	sat.mo = sat.mo * DEG2RAD

	var year int64 = 0
	if sat.epochyr < 57 {
		year = sat.epochyr + 2000
	} else {
		year = sat.epochyr + 1900
	}

	mon, day, hr, min, sec := days2mdhms(year, sat.epochdays)

	sat.jdsatepoch = JDay(int(year), int(mon), int(day), int(hr), int(min), int(sec))

	sgp4init(&opsmode, sat.jdsatepoch-2433281.5, sat)

	return sat, nil
}

// Parses a string into a float64 value.
func parseFloat(strIn string) (float64, error) {
	return strconv.ParseFloat(strIn, 64)
}

// Parses a string into a int64 value.
func parseInt(strIn string) (int64, error) {
	return strconv.ParseInt(strIn, 10, 0)
}
