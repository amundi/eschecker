package main

import (
	"github.com/amundi/escheck/config"
	"github.com/amundi/escheck/eslog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Scheduler(t *testing.T) {
	sched := new(scheduler)
	eslog.InitSilent()
	info := &config.Query{
		Schedule:       "500s",
		Alert_onlyonce: true,
		Alert_endmsg:   true,
	}

	sched.initScheduler(info)
	assert.Equal(t, true, sched.isAlertOnlyOnce, "Should be true")
	assert.Equal(t, true, sched.isAlertEndMsg, "Should be true")
	assert.Equal(t, "8m20s", sched.waitSchedule.String(), "Should be equal")

	info = &config.Query{
		Schedule:       "30m",
		Alert_onlyonce: false,
		Alert_endmsg:   false,
	}
	sched.initScheduler(info)
	assert.Equal(t, false, sched.isAlertOnlyOnce, "Should be equal")
	assert.Equal(t, false, sched.isAlertEndMsg, "Should be false")
	assert.Equal(t, "30m0s", sched.waitSchedule.String(), "Should be equal")

	info = &config.Query{
		Alert_onlyonce: false,
	}
	sched.initScheduler(info)
	assert.Equal(t, true, sched.isAlertOnlyOnce, "Should be equal")
	assert.Equal(t, false, sched.isAlertEndMsg, "Should be false")
	assert.Equal(t, "10m0s", sched.waitSchedule.String(), "Should be equal")
	assert.Equal(t, "10m0s", sched.alertSchedule.String(), "Should be equal")

	info = &config.Query{}
	sched.initScheduler(info)
	assert.Equal(t, true, sched.isAlertOnlyOnce, "Should be equal")
	assert.Equal(t, false, sched.isAlertEndMsg, "Should be false")
	assert.Equal(t, "10m0s", sched.waitSchedule.String(), "Should be equal")
	assert.Equal(t, "10m0s", sched.alertSchedule.String(), "Should be equal")

	info = &config.Query{
		Schedule:       "pouet",
		Alert_onlyonce: false,
	}
	sched.initScheduler(info)
	assert.Equal(t, true, sched.isAlertOnlyOnce, "Should be equal")
	assert.Equal(t, false, sched.isAlertEndMsg, "Should be false")
	assert.Equal(t, "10m0s", sched.waitSchedule.String(), "Should be equal")
	assert.Equal(t, "10m0s", sched.alertSchedule.String(), "Should be equal")

	info = &config.Query{
		Schedule:       "40z",
		Alert_onlyonce: false,
	}
	assert.Equal(t, true, sched.isAlertOnlyOnce, "Should be equal")
	assert.Equal(t, false, sched.isAlertEndMsg, "Should be false")
	assert.Equal(t, "10m0s", sched.waitSchedule.String(), "Should be equal")
	assert.Equal(t, "10m0s", sched.alertSchedule.String(), "Should be equal")
	sched.initScheduler(info)
}
